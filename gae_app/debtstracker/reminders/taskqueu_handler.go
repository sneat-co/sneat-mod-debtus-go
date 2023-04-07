package reminders

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal/gaedal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
	"github.com/strongo/nds"
)

func SendReminderHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	//func sendNotificationForDueTransfer(c context.Context, key *datastore.Key) {
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c, "Failed to parse form")
		return
	}
	reminderID, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		log.Errorf(c, "Failed to convert reminder ID to int")
		return
	}
	if err = sendReminder(c, reminderID); err != nil {
		log.Errorf(c, err.Error())
		if !db.IsNotFound(err) {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func sendReminder(c context.Context, reminderID int64) error {
	log.Debugf(c, "sendReminder(reminderID=%v)", reminderID)
	if reminderID == 0 {
		return errors.New("reminderID == 0")
	}

	reminder, err := dtdal.Reminder.GetReminderByID(c, reminderID)
	if err != nil {
		return err
	}
	if reminder.Status != models.ReminderStatusCreated {
		log.Infof(c, "reminder.Status:%v != models.ReminderStatusCreated", reminder.Status)
		return nil
	}

	transfer, err := facade.Transfers.GetTransferByID(c, reminder.TransferID)
	if err != nil {
		if db.IsNotFound(err) {
			log.Errorf(c, err.Error())
			if err = dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
				if reminder, err = dtdal.Reminder.GetReminderByID(c, reminderID); err != nil {
					return
				}
				reminder.Status = "invalid:no-transfer"
				reminder.DtUpdated = time.Now()
				reminder.DtNext = time.Time{}
				if err = dtdal.Reminder.SaveReminder(c, reminder); err != nil {
					return
				}
				return
			}, dtdal.SingleGroupTransaction); err != nil {
				return errors.Wrap(err, "Failed to update reminder")
			}
			return nil
		} else {
			return errors.Wrap(err, "Failed to load transfer")
		}
	}

	if !transfer.IsOutstanding {
		log.Infof(c, "Transfer(id=%v) is not outstanding, transfer.Amount=%v, transfer.AmountInCentsReturned=%v", reminder.TransferID, transfer.AmountInCents, transfer.AmountReturned())
		if err := gaedal.DiscardReminder(c, reminderID, reminder.TransferID, 0); err != nil {
			return errors.Wrapf(err, "Failed to discard a reminder for non outstanding transfer id=%v", reminder.TransferID)
		}
		return nil
	}

	if err = sendReminderToUser(c, reminderID, transfer); err != nil {
		log.Errorf(c, "Failed to send reminder (id=%v) for transfer %v: %v", reminderID, reminder.TransferID, err.Error())
	}

	return nil
}

var errReminderAlreadySentOrIsBeingSent = errors.New("Reminder already sent or is being sent")

func sendReminderToUser(c context.Context, reminderID int64, transfer models.Transfer) (err error) {

	var reminder models.Reminder

	// If sending notification failed do not try to resend - to prevent spamming.
	if err = dtdal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		reminder, err = dtdal.Reminder.GetReminderByID(c, reminderID)

		if reminder, err = dtdal.Reminder.GetReminderByID(c, reminderID); err != nil {
			return errors.Wrapf(err, "Failed to get reminder by id=%v", reminderID)
		}
		if reminder.Status != models.ReminderStatusCreated {
			return errReminderAlreadySentOrIsBeingSent
		}
		reminder.Status = models.ReminderStatusSending
		if err = dtdal.Reminder.SaveReminder(tc, reminder); err != nil { // TODO: User dtdal.Reminder.SaveReminder()
			return errors.Wrap(err, "Failed to save reminder with new status to db")
		}
		return
	}, nil); err != nil {
		if err == errReminderAlreadySentOrIsBeingSent {
			log.Infof(c, err.Error())
		} else {
			err = errors.WithMessage(err, fmt.Sprintf("failed to update reminder status to '%v'", models.ReminderStatusSending))
			log.Errorf(c, err.Error())
		}
		return
	} else {
		log.Infof(c, "Updated Reminder(id=%v) status to '%v'.", reminderID, models.ReminderStatusSending)
	}

	var user models.AppUserEntity
	if err = nds.Get(c, gaedal.NewAppUserKey(c, reminder.UserID), &user); err != nil {
		err = errors.Wrapf(err, "Failed to get user by id=%v", transfer.CreatorUserID)
		return
	}

	var reminderIsSent, channelDisabledByUser bool
	if user.HasTelegramAccount() {
		var (
			tgChatID int64
			tgBotID  string
		)
		if transferUserInfo := transfer.UserInfoByUserID(reminder.UserID); transferUserInfo.TgChatID != 0 {
			tgChatID = transferUserInfo.TgChatID
			tgBotID = transferUserInfo.TgBotID
		} else {
			var tgChat *telegram.TgChatEntityBase
			_, tgChat, err = gaedal.GetTelegramChatByUserID(c, reminder.UserID) // TODO: replace with DAL method
			if err != nil {
				if db.IsNotFound(err) { // TODO: Get rid of datastore reference
					err = errors.WithMessage(err, fmt.Sprintf("failed to call gaedal.GetTelegramChatByUserID(userID=%v)", reminder.UserID))
					return
				}
			} else {
				tgChatID = (int64)(tgChat.TelegramUserID)
				tgBotID = tgChat.BotID
			}
		}
		if tgChatID != 0 {
			if reminderIsSent, channelDisabledByUser, err = sendReminderByTelegram(c, transfer, reminder, tgChatID, tgBotID); err != nil {
				return
			} else if !reminderIsSent && !channelDisabledByUser {
				log.Warningf(c, "Reminder is not sent to Telegram, err=%v", err)
			}
		}
	}
	if !reminderIsSent { // TODO: This is wrong to send same reminder by email if Telegram failed, complex and will screw up stats <= Are you sure?
		if user.EmailAddress != "" {
			if err = sendReminderByEmail(c, reminder, user.EmailAddress, transfer, user); err != nil {
				log.Errorf(c, "Failure in sendReminderByEmail()")
			}
		} else {
			if !channelDisabledByUser {
				log.Errorf(c, "Can't send reminder")
			}
			err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
				if reminder, err = dtdal.Reminder.GetReminderByID(c, reminderID); err != nil {
					return err
				}
				reminder.Status = models.ReminderStatusFailed
				return dtdal.Reminder.SaveReminder(c, reminder)
			}, nil)
			if err != nil {
				log.Errorf(c, errors.WithMessage(err, fmt.Sprintf("failed to set reminder status to '%v'", models.ReminderStatusFailed)).Error())
			} else {
				log.Infof(c, "Reminder status set to '%v'", reminder.Status)
			}
		}
	}
	return nil // TODO: Handle errors!
}
