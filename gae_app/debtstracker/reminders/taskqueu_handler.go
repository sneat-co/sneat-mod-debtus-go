package reminders

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal/gaedal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/log"
	"github.com/strongo/nds"
	"net/http"
	"strconv"
	"time"
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
		if !dal.IsNotFound(err) {
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

	transfer, err := facade.Transfers.GetTransferByID(c, tx, reminder.TransferID)
	if err != nil {
		if dal.IsNotFound(err) {
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
				return fmt.Errorf("failed to update reminder: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to load transfer: %w", err)
		}
	}

	if !transfer.Data.IsOutstanding {
		log.Infof(c, "Transfer(id=%v) is not outstanding, transfer.Amount=%v, transfer.AmountInCentsReturned=%v", reminder.TransferID, transfer.Data.AmountInCents, transfer.Data.AmountReturned())
		if err := gaedal.DiscardReminder(c, reminderID, reminder.TransferID, 0); err != nil {
			return fmt.Errorf("failed to discard a reminder for non outstanding transfer id=%v: %w", reminder.TransferID, err)
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
			return fmt.Errorf("failed to get reminder by id=%v: %w", reminderID, err)
		}
		if reminder.Status != models.ReminderStatusCreated {
			return errReminderAlreadySentOrIsBeingSent
		}
		reminder.Status = models.ReminderStatusSending
		if err = dtdal.Reminder.SaveReminder(tc, reminder); err != nil { // TODO: User dtdal.Reminder.SaveReminder()
			return fmt.Errorf("failed to save reminder with new status to db: %w", err)
		}
		return
	}, nil); err != nil {
		if err == errReminderAlreadySentOrIsBeingSent {
			log.Infof(c, err.Error())
		} else {
			err = fmt.Errorf("failed to update reminder status to '%v': %w", models.ReminderStatusSending, err)
			log.Errorf(c, err.Error())
		}
		return
	} else {
		log.Infof(c, "Updated Reminder(id=%v) status to '%v'.", reminderID, models.ReminderStatusSending)
	}

	var user models.AppUserEntity
	if err = nds.Get(c, gaedal.NewAppUserKey(c, reminder.UserID), &user); err != nil {
		err = fmt.Errorf("failed to get user by id=%v: %w", transfer.Data.CreatorUserID, err)
		return
	}

	var reminderIsSent, channelDisabledByUser bool
	if user.HasTelegramAccount() {
		var (
			tgChatID int64
			tgBotID  string
		)
		if transferUserInfo := transfer.Data.UserInfoByUserID(reminder.UserID); transferUserInfo.TgChatID != 0 {
			tgChatID = transferUserInfo.TgChatID
			tgBotID = transferUserInfo.TgBotID
		} else {
			var tgChat *telegram.TgChatEntityBase
			_, tgChat, err = gaedal.GetTelegramChatByUserID(c, reminder.UserID) // TODO: replace with DAL method
			if err != nil {
				if dal.IsNotFound(err) { // TODO: Get rid of datastore reference
					err = fmt.Errorf("failed to call gaedal.GetTelegramChatByUserID(userID=%v): %w", reminder.UserID, err)
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
				log.Errorf(c, fmt.Errorf("failed to set reminder status to '%v': %w", models.ReminderStatusFailed, err).Error())
			} else {
				log.Infof(c, "Reminder status set to '%v'", reminder.Status)
			}
		}
	}
	return nil // TODO: Handle errors!
}
