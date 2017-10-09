package reminders

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal/gaedal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/qedus/nds"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"time"
	"github.com/strongo/app/db"
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

	reminder, err := dal.Reminder.GetReminderByID(c, reminderID)
	if err != nil {
		return err
	}
	if reminder.Status != models.ReminderStatusCreated {
		log.Infof(c, "reminder.Status:%v != models.ReminderStatusCreated", reminder.Status)
		return nil
	}

	transfer, err := dal.Transfer.GetTransferByID(c, reminder.TransferID)
	if err != nil {
		if db.IsNotFound(err) {
			log.Errorf(c, err.Error())
			if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
				if reminder, err = dal.Reminder.GetReminderByID(c, reminderID); err != nil {
					return
				}
				reminder.Status = "invalid:no-transfer"
				reminder.DtUpdated = time.Now()
				reminder.DtNext = time.Time{}
				if err = dal.Reminder.SaveReminder(c, reminder); err != nil {
					return
				}
				return
			}, dal.SingleGroupTransaction); err != nil {
				return errors.Wrap(err, "Failed to update reminder")
			}
			return nil
		} else {
			return errors.Wrap(err, "Failed to load transfer")
		}
	}

	if !transfer.IsOutstanding {
		log.Infof(c, "Transfer(id=%v) is not outstanding, transfer.Amount=%v, transfer.AmountReturned=%v", reminder.TransferID, transfer.AmountInCents, transfer.AmountInCentsReturned)
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
	if err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		reminder, err = dal.Reminder.GetReminderByID(c, reminderID)

		if reminder, err = dal.Reminder.GetReminderByID(c, reminderID); err != nil {
			return errors.Wrapf(err, "Failed to get reminder by id=%v", reminderID)
		}
		if reminder.Status != models.ReminderStatusCreated {
			return errReminderAlreadySentOrIsBeingSent
		}
		reminder.Status = models.ReminderStatusSending
		if _, err := nds.Put(tc, gaedal.NewReminderKey(tc, reminderID), reminder.ReminderEntity); err != nil { // TODO: User dal.Reminder.SaveReminder()
			return errors.Wrap(err, "Failed to save reminder with new status to db")
		}
		return nil
	}, nil); err != nil {
		if err == errReminderAlreadySentOrIsBeingSent {
			log.Infof(c, err.Error())
		} else {
			log.Errorf(c, "Failed to update reminder status to '%v': %v", models.ReminderStatusSending, err)
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
		if reminder.UserID == transfer.CreatorUserID && transfer.CreatorTgChatID != 0 {
			tgChatID = transfer.CreatorTgChatID
			tgBotID = transfer.CreatorTgBotID
		} else {
			//user.TelegramUserIDs[0]
			var tgChat *telegram_bot.TelegramChatEntityBase
			_, tgChat, err = gaedal.GetTelegramChatByUserID(c, reminder.UserID)
			if err != nil {
				err = errors.Wrap(err, "Failed to GetTelegramChatByUserID() ")
			}
			tgChatID = (int64)(tgChat.TelegramUserID)
			tgBotID = tgChat.BotID
		}
		reminderIsSent, channelDisabledByUser, err = sendReminderByTelegram(c, TransferReminderToCreator, transfer, reminderID, reminder.UserID, tgChatID, tgBotID)
		if err != nil {
			return
		}
		if !reminderIsSent && !channelDisabledByUser {
			log.Warningf(c, "Reminder is not sent to Telegram")
		}
	}
	if !reminderIsSent { // TODO: This is wrong to send same reminder by email if Telegram failed, complex and will screw up stats
		if user.EmailAddress != "" {
			if err = sendReminderByEmail(c, reminder, user.EmailAddress, transfer, user); err != nil {
				log.Errorf(c, "Failure in sendReminderByEmail()")
			}
		} else {
			if !channelDisabledByUser {
				log.Errorf(c, "Can't send reminder")
			}
			err = dal.DB.RunInTransaction(c, func(c context.Context) error {
				reminderKey := gaedal.NewReminderKey(c, reminderID)
				if err = nds.Get(c, reminderKey, reminder.ReminderEntity); err != nil {
					return err
				}
				reminder.Status = models.ReminderStatusFailed
				_, err = nds.Put(c, reminderKey, reminder.ReminderEntity)
				return err
			}, nil)
			if err != nil {
				log.Errorf(c, errors.Wrapf(err, "Failed to set reminder status to '%v'", models.ReminderStatusFailed).Error())
			} else {
				log.Infof(c, "Reminder status set to '%v'", models.ReminderStatusFailed)
			}
		}
	}
	return nil // TODO: Handle errors!
}
