package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/app/gae"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
	"strconv"
	"strings"
	"time"
	"github.com/strongo/app/gaedb"
	"bitbucket.com/asterus/debtstracker-server/gae_app/gaestandard"
)

func _delayReminderCreation(c context.Context, transferID int64, f *delay.Function) error {
	if transferID == 0 {
		panic("transferID == 0")
	}
	if task, err := gae.CreateDelayTask(common.QUEUE_REMINDERS, "create-reminder", f, transferID); err != nil {
		return errors.Wrapf(err, "Failed to create a task for reminder creation, transfer id=%v", transferID)
	} else {
		task.Delay = time.Duration(time.Second)
		if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			return errors.Wrapf(err, "Failed to add a task for reminder creation, transfer id=%v", transferID)
		}
		log.Debugf(c, "Added task(%v) to create reminder for transfer id=%v", f, transferID)
	}
	return nil
}

func (_ ReminderDalGae) DelayCreateReminderForTransferCreator(c context.Context, transferID int64) error {
	log.Debugf(c, "DelayCreateReminderForTransferCreator(transferID=%v)", transferID)
	return _delayReminderCreation(c, transferID, _delayedCreateReminderForTransferCreator)
}

func (_ ReminderDalGae) DelayCreateReminderForTransferCounterparty(c context.Context, transferID int64) error {
	return _delayReminderCreation(c, transferID, _delayedCreateReminderForTransferCounterparty)
}

var _delayedCreateReminderForTransferCreator = delay.Func("_createReminderForTransferCreator", _createReminderForTransferCreator)
var _delayedCreateReminderForTransferCounterparty = delay.Func("_createReminderForTransferCounterparty", _createReminderForTransferCounterparty)

//type remindersFactory struct {
//
//}
//var _remindersFactory = remindersFactory{}
//
//func (f remindersFactory) createReminderForCounterparty(c context.Context, transferID int64) {
//
//}

func _createReminderForTransferCounterparty(c context.Context, transferID int64) error {
	log.Debugf(c, "_createReminderForTransferCounterparty(transferID=%v)", transferID)
	if transferID == 0 {
		log.Errorf(c, "transferID == 0")
		return nil
	}
	transfer, err := dal.Transfer.GetTransferByID(c, transferID)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(c, "Transfer not found by id=%v", transferID)
			return nil
		}
		return errors.Wrap(err, "Failed to get transfer by id")
	}

	var (
		counterpartyTgChatID int64
		counterpartyTgChat   *telegram_bot.TelegramChatEntityBase
		//counterpartyUser models.AppUser
	)
	if transfer.CounterpartyReminderID == 0 && transfer.Counterparty().UserID != 0 {
		var entityID string
		entityID, counterpartyTgChat, err = GetTelegramChatByUserID(c, transfer.Counterparty().UserID)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				log.Infof(c, "User transfer.Contact().UserID=%v has no Telegram chat", transfer.Counterparty().UserID)
				return nil
			} else {
				return errors.Wrapf(err, "Failed to get Telegram chat for counterparty user (id=%v)", transfer.Counterparty().UserID)
			}
		}
		if entityID == "" {
			log.Infof(c, "User transfer.Contact().UserID=%v has no Telegram chat", transfer.Counterparty().UserID)
			return nil
		}

		var counterpartyTgChatStringID string
		if strings.Contains(entityID, ":") { // TODO: Temporary workaround for migrating from Int to Str ids.
			counterpartyTgChatStringID = strings.Split(entityID, ":")[1]
		} else {
			counterpartyTgChatStringID = entityID
		}
		if counterpartyTgChatID, err = strconv.ParseInt(counterpartyTgChatStringID, 10, 64); err != nil {
			return errors.Wrapf(err, "Failed to strconv.ParseInt(strings.Split(entityID, ':')[1], 10, 64): [%v]", entityID)
		}
	}

	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		transfer, err = dal.Transfer.GetTransferByID(c, transferID)
		if err != nil {
			return errors.Wrap(err, "Failed to get transfer by id in transaction")
		}
		isTransferChanged := false
		if transfer.CounterpartyReminderID == 0 {
			if transfer.CounterpartyTgChatID == 0 && counterpartyTgChatID != 0 {
				transfer.CounterpartyTgChatID = counterpartyTgChatID
				transfer.CounterpartyTgBotID = counterpartyTgChat.BotID
				isTransferChanged = true
			}
			if transfer.CounterpartyTgChatID != 0 {
				reminderKey := NewReminderIncompleteKey(c)
				reminder := models.NewReminderViaTelegram(transfer.CounterpartyTgBotID, transfer.CounterpartyTgChatID, transfer.Counterparty().UserID, transferID, false, transfer.DtDueOn)
				if reminderKey, err = gaedb.Put(c, reminderKey, &reminder); err != nil {
					return errors.Wrap(err, "Faield to save reminder to datastore")
				}
				log.Infof(c, "Created reminder id=%v", reminderKey.IntID())
				dueIn := transfer.DtDueOn.Sub(time.Now())
				if err = QueueSendReminder(c, reminderKey.IntID(), dueIn); err != nil {
					return errors.Wrap(err, "Failed to queue reminder for sending")
				}
				transfer.CreatorReminderID = reminderKey.IntID()
				isTransferChanged = true
			} else {
				log.Infof(c, "Can not send reminder as transfer.CounterpartyTgChatID == 0")
			}
			return nil
		}
		if isTransferChanged {
			if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
				err = errors.Wrap(err, "Failed to save transfer")
			}
		}
		return err
	}, dal.CrossGroupTransaction)
	return nil
}

func _createReminderForTransferCreator(c context.Context, transferID int64) error {
	log.Debugf(c, "_createReminderForTransferCreator(transferID=%v)", transferID)
	if transferID == 0 {
		log.Errorf(c, "transferID == 0")
		return nil
	}
	//transferKey, transfer, err := transferDal.GetTransferByID(c, transferID)
	//if err != nil {
	//	return errors.Wrap(err, "Failed to get transfer by id")
	//}

	err := dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		transfer, err := dal.Transfer.GetTransferByID(c, transferID)
		if err != nil {
			return errors.Wrap(err, "Failed to get transfer by id in transaction")
		}
		isTransferChanged := false
		if transfer.CreatorReminderID == 0 {
			if transfer.CreatorTgChatID != 0 {
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
				reminder := models.NewReminderViaTelegram(transfer.CreatorTgBotID, transfer.CreatorTgChatID, transfer.CreatorUserID, transferID, isAutomatic, next)
				if reminderKey, err = gaedb.Put(c, reminderKey, &reminder); err != nil {
					return errors.Wrap(err, "Failed to save reminder to datastore")
				}
				log.Infof(c, "Created reminder id=%v", reminderKey.IntID())
				if err = QueueSendReminder(c, reminderKey.IntID(), next.Sub(time.Now())); err != nil {
					return errors.Wrap(err, "Failed to queue reminder for sending")
				}
				transfer.CreatorReminderID = reminderKey.IntID()
				isTransferChanged = true
			}
		}
		if isTransferChanged {
			if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
				return errors.Wrap(err, "Failed to save transfer to datastore")
			}
		}
		return err
	}, dal.CrossGroupTransaction)
	return err
}

func (_ ReminderDalGae) DelayDiscardReminders(c context.Context, transferIDs []int64, returnTransferID int64) error {
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
	if err := _discard(dal.Reminder.GetActiveReminderIDsByTransferID, "Loaded %v keys of active reminders for transfer id=%v", "The are no ative reminders for transfer id=%v"); err != nil {
		return err
	}
	if err := _discard(dal.Reminder.GetSentReminderIDsByTransferID, "Loaded %v keys of sent reminders for transfer id=%v", "The are no sent reminders for transfer id=%v"); err != nil {
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
		reminder       = models.Reminder{ID: reminderID, ReminderEntity: new(models.ReminderEntity)}
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

	if reminder, err = dal.Reminder.SetReminderStatus(c, reminderID, returnTransferID, models.ReminderStatusDiscarded, time.Now()); err != nil {
		return err // DO NOT WRAP as there is check in delayedDiscardReminder() errors.Wrapf(err, "Failed to set reminder status to '%v'", models.ReminderStatusDiscarded)
	}

	switch reminder.SentVia {
	case telegram_bot.TelegramPlatformID: // We need to update a reminder message if it was already sent out
		if reminder.BotID == "" {
			log.Errorf(c, "reminder.BotID == ''")
			return nil
		}
		if reminder.MessageIntID == 0 {
			//log.Infof(c, "No need to update reminder message in Telegram as a reminder is not sent yet")
			return nil
		}
		log.Infof(c, "Will try to update a reminder message as it was already sent to user, reminder.MessageIntID: %v", reminder.MessageIntID)
		tgBotApi := telegram.GetTelegramBotApiByBotCode(c, reminder.BotID)
		if tgBotApi == nil {
			return errors.New(fmt.Sprintf("Not able to create API client as there no settings for telegram bot with id '%v'", reminder.BotID))
		}

		if reminder.Locale == "" {
			log.Errorf(c, "reminder.Locale == ''")
			if user, err := dal.User.GetUserByID(c, reminder.UserID); err != nil {
				return errors.Wrapf(err, "Failed to get user by id=%v", reminder.UserID)
			} else if user.PreferredLanguage != "" {
				reminder.Locale = user.PreferredLanguage
			} else if s, ok := telegram.Bots(gaestandard.GetEnvironment(c), nil).ByCode[reminder.BotID]; ok {
				reminder.Locale = s.Locale.Code5
			}
		}

		executionContext := GetExecutionContextForReminder(c, reminder.ReminderEntity)

		utmParams := common.UtmParams{
			Source:   "TODO", // TODO: Get bot ID
			Medium:   telegram_bot.TelegramPlatformID,
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

func (_ ReminderDalGae) SetReminderStatus(c context.Context, reminderID, returnTransferID int64, status string, when time.Time) (reminder models.Reminder, err error) {
	var (
		changed        bool
		previousStatus string
	)
	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if reminder, err = dal.Reminder.GetReminderByID(c, reminderID); err != nil {
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
