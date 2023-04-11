package gaedal

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/strongo/app"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/gaestandard"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
)

func (ReminderDalGae) DelayCreateReminderForTransferUser(c context.Context, transferID, userID int64) (err error) {
	if transferID == 0 {
		panic("transferID == 0")
	}
	if userID == 0 {
		panic("userID == 0")
	}
	if !dtdal.DB.IsInTransaction(c) {
		panic("This function should be called within transaction")
	}
	if task, err := gae.CreateDelayTask(common.QUEUE_REMINDERS, "create-reminder-4-transfer-user", delayCreateReminderForTransferUser, transferID, userID); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("Failed to create a task for reminder creation. transferID=%v, userID=%v", transferID, userID))
	} else {
		task.Delay = time.Duration(time.Second)
		if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("Failed to add a task for reminder creation, transfer id=%v", transferID))
		}
		log.Debugf(c, "Added task(%v) to create reminder for transfer id=%v", task.Path, transferID)
	}
	return
}

var delayCreateReminderForTransferUser = delay.Func("createReminderForTransferUser", createReminderForTransferUser)

func createReminderForTransferUser(c context.Context, transferID, userID int64) (err error) {
	log.Debugf(c, "createReminderForTransferUser(transferID=%d, userID=%d)", transferID, userID)
	if transferID == 0 {
		log.Errorf(c, "transferID == 0")
		return nil
	}
	if userID == 0 {
		log.Errorf(c, "userID == 0")
		return nil
	}

	return dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var transfer models.Transfer
		transfer, err = facade.Transfers.GetTransferByID(c, transferID)
		if err != nil {
			if dal.IsNotFound(err) {
				log.Errorf(c, errors.WithMessage(err, "Not able to create reminder for specified transfer").Error())
				return
			}
			return errors.WithMessage(err, "failed to get transfer by id")
		}
		transferUserInfo := transfer.UserInfoByUserID(userID)
		if transferUserInfo.UserID != userID {
			panic("transferUserInfo.UserID != userID")
		}

		if transferUserInfo.ReminderID != 0 {
			log.Warningf(c, "Transfer user already has reminder # %v", transferUserInfo.ReminderID)
			return
		}

		if transferUserInfo.TgChatID == 0 { // TODO: Try to get TgChat from user record or check other channels?
			log.Warningf(c, "Transfer user has no associated TgChatID: %+v", transferUserInfo)
			return
		}

		reminderKey := NewReminderIncompleteKey(c)
		next := transfer.DtDueOn
		isAutomatic := next.IsZero()
		if isAutomatic {
			if strings.Contains(strings.ToLower(transfer.CreatedOnID), "dev") {
				next = time.Now().Add(2 * time.Minute)
			} else {
				next = time.Now().Add(7 * 24 * time.Hour)
			}
		}
		reminder := models.NewReminderViaTelegram(transferUserInfo.TgBotID, transferUserInfo.TgChatID, userID, transferID, isAutomatic, next)
		if reminderKey, err = gaedb.Put(c, reminderKey, &reminder); err != nil {
			return errors.WithMessage(err, "failed to save reminder to datastore")
		}
		log.Infof(c, "Created reminder id=%v", reminderKey.IntID())
		if err = QueueSendReminder(c, reminderKey.IntID(), next.Sub(time.Now())); err != nil {
			return errors.WithMessage(err, "failed to queue reminder for sending")
		}
		transferUserInfo.ReminderID = reminderKey.IntID()

		if err = facade.Transfers.SaveTransfer(c, transfer); err != nil {
			return errors.WithMessage(err, "failed to save transfer to datastore")
		}

		return
	}, dtdal.CrossGroupTransaction)
}

func (ReminderDalGae) DelayDiscardReminders(c context.Context, transferIDs []int64, returnTransferID int64) error {
	if len(transferIDs) > 0 {
		return gae.CallDelayFunc(c, common.QUEUE_REMINDERS, "discard-reminders", delayDiscardReminders, transferIDs, returnTransferID)
	} else {
		log.Warningf(c, "DelayDiscardReminders(): len(transferIDs)==0")
		return nil
	}
}

var delayDiscardReminders = delay.Func("discardReminders", discardReminders)

func discardReminders(c context.Context, transferIDs []int64, returnTransferID int64) error {
	log.Debugf(c, "discardReminders(transferIDs=%v, returnTransferID=%returnTransferID)", transferIDs, returnTransferID)
	if len(transferIDs) == 0 {
		return errors.New("len(transferIDs) == 0")
	}
	const queueName = common.QUEUE_REMINDERS
	tasks := make([]*taskqueue.Task, len(transferIDs))
	for i, transferID := range transferIDs {
		if task, err := gae.CreateDelayTask(queueName, "discard-reminders-for-transfer", delayDiscardRemindersForTransfer, transferID, returnTransferID); err != nil {
			return errors.Wrapf(err, "Failed to create delay task to dicard reminder for transfer id=%v", transferID)
		} else {
			tasks[i] = task
		}
	}
	if _, err := taskqueue.AddMulti(c, tasks, queueName); err != nil {
		return errors.Wrapf(err, "Failed to add %v task(s) to queue '%v'", len(tasks), queueName)
	}
	return nil
}

var delayDiscardRemindersForTransfer = delay.Func("discardRemindersForTransfer", discardRemindersForTransfer)

func discardRemindersForTransfer(c context.Context, transferID, returnTransferID int64) error {
	log.Debugf(c, "discardReminders(transferID=%v, returnTransferID=%v)", transferID, returnTransferID)
	if transferID == 0 {
		log.Errorf(c, "transferID == 0")
		return nil
	}
	var tasks []*taskqueue.Task
	delayDuration := time.Millisecond * 10
	var _discard = func(
		getIDs func(context.Context, int64) ([]int64, error),
		loadedFormat, notLoadedFormat string,
	) error {
		if reminderIDs, err := getIDs(c, transferID); err != nil {
			return err
		} else if len(reminderIDs) > 0 {
			log.Debugf(c, loadedFormat, len(reminderIDs), transferID)
			for _, reminderID := range reminderIDs {
				if task, err := gae.CreateDelayTask(common.QUEUE_REMINDERS, "discard-reminder", delayDiscardReminder, reminderID, transferID, returnTransferID); err != nil {
					return errors.Wrapf(err, "Failed to create a task for reminder %v", reminderID)
				} else {
					task.Delay = delayDuration
					tasks = append(tasks, task)
					delayDuration += time.Millisecond * 10
				}
			}
		} else {
			log.Infof(c, notLoadedFormat, transferID)
		}
		return nil
	}
	if err := _discard(dtdal.Reminder.GetActiveReminderIDsByTransferID, "Loaded %v keys of active reminders for transfer id=%v", "The are no ative reminders for transfer id=%v"); err != nil {
		return err
	}
	if err := _discard(dtdal.Reminder.GetSentReminderIDsByTransferID, "Loaded %v keys of sent reminders for transfer id=%v", "The are no sent reminders for transfer id=%v"); err != nil {
		return err
	}
	if len(tasks) > 0 {
		if _, err := taskqueue.AddMulti(c, tasks, common.QUEUE_REMINDERS); err != nil {
			return errors.Wrapf(err, "Failed to put %v tasks to queue", len(tasks))
		}
	}
	return nil
}

var delayDiscardReminder = delay.Func("DiscardReminder", delayedDiscardReminder)

func DiscardReminder(c context.Context, reminderID, transferID, returnTransferID int64) (err error) {
	return discardReminder(c, reminderID, transferID, returnTransferID)
}

func delayedDiscardReminder(c context.Context, reminderID, transferID, returnTransferID int64) (err error) {
	if discardReminder(c, reminderID, transferID, returnTransferID); err == ErrDuplicateAttemptToDiscardReminder {
		log.Errorf(c, err.Error())
		return nil
	}
	return err
}

func discardReminder(c context.Context, reminderID, transferID, returnTransferID int64) (err error) {
	log.Debugf(c, "discardReminder(reminderID=%v, transferID=%v, returnTransferID=%v)", reminderID, transferID, returnTransferID)

	reminderKey := NewReminderKey(c, reminderID)
	transferKey := NewTransferKey(c, transferID)

	var (
		transferEntity models.TransferEntity
		reminder       = models.NewReminder(reminderID, new(models.ReminderEntity))
	)

	if returnTransferID > 0 {
		returnTransferKey := NewTransferKey(c, returnTransferID)
		var returnTransfer models.TransferEntity
		keys := []*datastore.Key{reminderKey, transferKey, returnTransferKey}
		if err = gaedb.GetMulti(c, keys, []interface{}{reminder.ReminderEntity, &transferEntity, &returnTransfer}); err != nil {
			return errors.Wrapf(err, "Failed to get entities from datastore by keys=%v", keys)
		}
	} else {
		keys := []*datastore.Key{reminderKey, transferKey}
		if err = gaedb.GetMulti(c, keys, []interface{}{reminder.ReminderEntity, &transferEntity}); err != nil {
			return errors.Wrapf(err, "Failed to get entities from datastore by keys=%v", keys)
		}
	}

	if reminder, err = dtdal.Reminder.SetReminderStatus(c, reminderID, returnTransferID, models.ReminderStatusDiscarded, time.Now()); err != nil {
		return err // DO NOT WRAP as there is check in delayedDiscardReminder() errors.Wrapf(err, "Failed to set reminder status to '%v'", models.ReminderStatusDiscarded)
	}

	switch reminder.SentVia {
	case telegram.PlatformID: // We need to update a reminder message if it was already sent out
		if reminder.BotID == "" {
			log.Errorf(c, "reminder.BotID == ''")
			return nil
		}
		if reminder.MessageIntID == 0 {
			//log.Infof(c, "No need to update reminder message in Telegram as a reminder is not sent yet")
			return nil
		}
		log.Infof(c, "Will try to update a reminder message as it was already sent to user, reminder.MessageIntID: %v", reminder.MessageIntID)
		tgBotApi := tgbots.GetTelegramBotApiByBotCode(c, reminder.BotID)
		if tgBotApi == nil {
			return fmt.Errorf("Not able to create API client as there no settings for telegram bot with id '%v'", reminder.BotID)
		}

		if reminder.Locale == "" {
			log.Errorf(c, "reminder.Locale == ''")
			if user, err := facade.User.GetUserByID(c, reminder.UserID); err != nil {
				return errors.Wrapf(err, "Failed to get user by id=%v", reminder.UserID)
			} else if user.PreferredLanguage != "" {
				reminder.Locale = user.PreferredLanguage
			} else if s, ok := tgbots.Bots(gaestandard.GetEnvironment(c), nil).ByCode[reminder.BotID]; ok {
				reminder.Locale = s.Locale.Code5
			}
		}

		executionContext := GetExecutionContextForReminder(c, reminder.ReminderEntity)

		utmParams := common.UtmParams{
			Source:   "TODO", // TODO: Get bot ID
			Medium:   telegram.PlatformID,
			Campaign: common.UTM_CAMPAIGN_RECEIPT_DISCARD,
		}

		receiptMessageText := common.TextReceiptForTransfer(
			executionContext,
			models.NewTransfer(transferID, &transferEntity),
			reminder.UserID,
			common.ShowReceiptToAutodetect,
			utmParams,
		)

		locale := strongo.GetLocaleByCode5(reminder.Locale) // TODO: Check for supported locales

		transferUrlForUser := common.GetTransferUrlForUser(transferID, reminder.UserID, locale, utmParams)

		receiptMessageText += "\n\n" + strings.Join([]string{
			executionContext.Translate(trans.MESSAGE_TEXT_DEBT_IS_RETURNED),
			fmt.Sprintf(`<a href="%v">%v</a>`, transferUrlForUser, executionContext.Translate(trans.MESSAGE_TEXT_DETAILS_ARE_HERE)),
		}, "\n")

		tgMessage := tgbotapi.NewEditMessageText(reminder.ChatIntID, int(reminder.MessageIntID), "", receiptMessageText)
		tgMessage.ParseMode = "HTML"
		if _, err = tgBotApi.Send(tgMessage); err != nil {
			return errors.Wrap(err, "Failed to send message to Telegram")
		}

	default:
		return errors.New("Unknown reminder channel: %v" + reminder.SentVia)
	}

	return err
}

func GetExecutionContextForReminder(c context.Context, reminder *models.ReminderEntity) strongo.ExecutionContext {
	translator := strongo.NewSingleMapTranslator(strongo.GetLocaleByCode5(reminder.Locale), strongo.NewMapTranslator(c, trans.TRANS))
	return strongo.NewExecutionContext(c, translator)
}

var ErrDuplicateAttemptToDiscardReminder = errors.New("Duplicate attempt to close reminder by same return transfer")

func (ReminderDalGae) SetReminderStatus(c context.Context, reminderID, returnTransferID int64, status string, when time.Time) (reminder models.Reminder, err error) {
	var (
		changed        bool
		previousStatus string
	)
	err = dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if reminder, err = dtdal.Reminder.GetReminderByID(c, reminderID); err != nil {
			return
		} else {
			switch status {
			case string(models.ReminderStatusDiscarded):
				reminder.DtDiscarded = when
			case string(models.ReminderStatusSent):
				reminder.DtSent = when
			case string(models.ReminderStatusSending):
				// pass
			case string(models.ReminderStatusViewed):
				reminder.DtViewed = when
			case string(models.ReminderStatusUsed):
				reminder.DtUsed = when
			default:
				return errors.New("Unsupported status: " + status)
			}
			previousStatus = reminder.Status
			changed = previousStatus != status
			if returnTransferID != 0 && status == string(models.ReminderStatusDiscarded) {
				for _, id := range reminder.ClosedByTransferIDs { // TODO: WTF are we doing here?
					if id == returnTransferID {
						log.Infof(c, "new status: '%v', Reminder{Status: '%v', ClosedByTransferIDs: %v}", status, reminder.Status, reminder.ClosedByTransferIDs)
						return ErrDuplicateAttemptToDiscardReminder
					}
				}
				reminder.ClosedByTransferIDs = append(reminder.ClosedByTransferIDs, returnTransferID)
				changed = true
			}
			if changed {
				reminder.Status = status
				if _, err = gaedb.Put(c, NewReminderKey(c, reminderID), reminder.ReminderEntity); err != nil {
					err = errors.Wrapf(err, "Failed to save reminder to db (id=%v)", reminderID)
				}
			}
			return
		}
	}, nil)
	if err == nil {
		if changed {
			log.Debugf(c, "Reminder(id=%v) status changed from '%v' to '%v'", reminderID, previousStatus, status)
		} else {
			log.Debugf(c, "Reminder(id=%v) status not changed as already '%v'", reminderID, status)
		}
	}
	return
}
