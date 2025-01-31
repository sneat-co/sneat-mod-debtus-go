package gaedal

import (
	"context"
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-core-modules/core/queues"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal/delayer4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/strongo/delaying"
	"github.com/strongo/logus"
	"reflect"
	"time"
)

func NewReminderIncompleteKey(_ context.Context) *dal.Key {
	return dal.NewIncompleteKey(models4debtus.ReminderKind, reflect.Int, nil)
}

func NewReminderKey(reminderID string) *dal.Key {
	if reminderID == "" {
		panic("reminderID == 0")
	}
	return dal.NewKeyWithID(models4debtus.ReminderKind, reminderID)
}

type ReminderDalGae struct {
}

func NewReminderDalGae() ReminderDalGae {
	return ReminderDalGae{}
}

var _ dtdal.ReminderDal = (*ReminderDalGae)(nil)

func (reminderDalGae ReminderDalGae) GetReminderByID(ctx context.Context, tx dal.ReadSession, id string) (reminder models4debtus.Reminder, err error) {
	reminder = models4debtus.NewReminder(id, nil)
	return reminder, tx.Get(ctx, reminder.Record)
}

func (reminderDalGae ReminderDalGae) SaveReminder(ctx context.Context, tx dal.ReadwriteTransaction, reminder models4debtus.Reminder) (err error) {
	return tx.Set(ctx, reminder.Record)
}

func (reminderDalGae ReminderDalGae) GetSentReminderIDsByTransferID(ctx context.Context, tx dal.ReadSession, transferID int) ([]int, error) {
	q := dal.From(dal.NewRootCollectionRef(models4debtus.ReminderKind, "")).Where(
		dal.WhereField("TransferID", dal.Equal, transferID),
		dal.WhereField("Status", dal.Equal, models4debtus.ReminderStatusSent),
	).SelectKeysOnly(reflect.Int)

	records, err := tx.QueryAllRecords(ctx, q)
	if err != nil {
		return nil, err
	}
	reminderIDs := make([]int, len(records))
	for i, record := range records {
		reminderIDs[i] = record.Key().ID.(int)
	}
	return reminderIDs, nil
}

func (reminderDalGae ReminderDalGae) GetActiveReminderIDsByTransferID(ctx context.Context, tx dal.ReadSession, transferID int) ([]int, error) {
	q := dal.From(dal.NewRootCollectionRef(models4debtus.ReminderKind, "")).Where(
		dal.WhereField("TransferID", dal.Equal, transferID),
		dal.WhereField("DtNext", dal.GreaterThen, time.Time{}),
	).SelectKeysOnly(reflect.Int)
	records, err := tx.QueryAllRecords(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to get active reminders by transfer id=%v: %w", transferID, err)
	}
	reminderIDs := make([]int, len(records))
	for i, record := range records {
		reminderIDs[i] = record.Key().ID.(int)
	}
	return reminderIDs, nil
}

func (reminderDalGae ReminderDalGae) SetReminderIsSent(ctx context.Context, reminderID string, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) (err error) {
	//gaehost.GaeLogger.Debugf(ctx, "delayedSetReminderIsSent(reminderID=%v, sentAt=%v, messageIntID=%v, messageStrID=%v)", reminderID, sentAt, messageIntID, messageStrID)
	if err := _validateSetReminderIsSentMessageIDs(messageIntID, messageStrID, sentAt); err != nil {
		return err
	}
	reminder := models4debtus.NewReminder(reminderID, nil)
	return facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return reminderDalGae.SetReminderIsSentInTransaction(ctx, tx, reminder, sentAt, messageIntID, messageStrID, locale, errDetails)
	})
}

func (reminderDalGae ReminderDalGae) SetReminderIsSentInTransaction(ctx context.Context, tx dal.ReadwriteTransaction, reminder models4debtus.Reminder, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) (err error) {
	if reminder.Data == nil {
		reminder, err = reminderDalGae.GetReminderByID(ctx, tx, reminder.ID)
		if err != nil {
			if dal.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("failed to get reminder by ContactID: %w", err)
		}
	}
	if reminder.Data.Status != models4debtus.ReminderStatusSending {
		logus.Errorf(ctx, "reminder.Status:%v != models.ReminderStatusSending:%v", reminder.Data.Status, models4debtus.ReminderStatusSending)
		return nil
	} else {
		reminder.Data.Status = models4debtus.ReminderStatusSent
		reminder.Data.DtSent = sentAt
		reminder.Data.DtScheduled = reminder.Data.DtNext
		reminder.Data.DtNext = time.Time{}
		reminder.Data.ErrDetails = errDetails
		reminder.Data.Locale = locale
		if messageIntID != 0 {
			reminder.Data.MessageIntID = messageIntID
		}
		if messageStrID != "" {
			reminder.Data.MessageStrID = messageStrID
		}
		if err = tx.Set(ctx, reminder.Record); err != nil {
			err = fmt.Errorf("failed to save reminder to datastore: %w", err)
		}
		return err
	}
}

func (reminderDalGae ReminderDalGae) RescheduleReminder(ctx context.Context, reminderID string, remindInDuration time.Duration) (oldReminder, newReminder models4debtus.Reminder, err error) {
	return models4debtus.Reminder{}, models4debtus.Reminder{}, errors.New("not implemented - needs to be refactored")
	//var (
	//	newReminderKey    *datastore.Key
	//	newReminderEntity *models.ReminderDbo
	//)
	//err = facade.RunReadwriteTransaction(ctx, func(tctx context.Context, tx dal.ReadwriteTransaction) (err error) {
	//	oldReminder, err = reminderDalGae.GetReminderByID(tctx, reminderID)
	//	if err != nil {
	//		return fmt.Errorf("failed to get oldReminder by id: %w", err)
	//	}
	//	if oldReminder.IsRescheduled {
	//		err = dtdal.ErrReminderAlreadyRescheduled
	//		return err
	//	}
	//	reminder := models.NewReminder(reminderID)
	//	if remindInDuration == time.Duration(0) {
	//		if _, err = tx.Set(tc, reminderKey, oldReminder.ReminderDbo); err != nil {
	//			return err
	//		}
	//	} else {
	//		nextReminderOn := time.Now().Add(remindInDuration)
	//		newReminderEntity = oldReminder.ScheduleNextReminder(reminderID, nextReminderOn)
	//		newReminderKey = NewReminderIncompleteKey(tc)
	//		keys, err := gaedb.PutMulti(tc, []*datastore.Key{reminderKey, newReminderKey}, []interface{}{oldReminder.ReminderDbo, newReminderEntity})
	//		if err != nil {
	//			err = fmt.Errorf("failed to reschedule oldReminder: %w", err)
	//		}
	//		newReminderKey = keys[1]
	//		if err = QueueSendReminder(tc, newReminderKey.IntID(), remindInDuration); err != nil { // TODO: Should be outside of DAL?
	//			return err
	//		}
	//	}
	//	return err
	//})
	//if err != nil {
	//	return
	//}
	//if newReminderKey != nil && newReminderEntity != nil {
	//	newReminder = models.Reminder{
	//		IntegerID:      db.NewIntID(newReminderKey.IntID()),
	//		ReminderDbo: newReminderEntity,
	//	}
	//}
	//return
}

func (ReminderDalGae) DelayCreateReminderForTransferUser(ctx context.Context, transferID string, userID string) (err error) {
	if transferID == "" {
		panic("transferID == 0")
	}
	if userID == "" {
		panic("userID == 0")
	}
	//if !dtdal.DB.IsInTransaction(ctx) {
	//	panic("This function should be called within transaction")
	//}
	if err = delayer4debtus.CreateReminderForTransferUser.EnqueueWork(ctx, delaying.With(queues.QueueReminders, "create-reminder-4-transfer-user", 0), transferID, userID); err != nil {
		return fmt.Errorf("failed to create a task for reminder creation. transferID=%v, userID=%v: %w", transferID, userID, err)
	}
	logus.Debugf(ctx, "Added task to create reminder for transfer id=%s", transferID)
	return
}
func (ReminderDalGae) DelayDiscardReminders(ctx context.Context, transferIDs []string, returnTransferID string) error {
	if len(transferIDs) > 0 {
		return delayer4debtus.DiscardReminders.EnqueueWork(ctx, delaying.With(queues.QueueReminders, "discard-reminders", 0), transferIDs, returnTransferID)
	} else {
		logus.Warningf(ctx, "DelayDiscardReminders(): len(transferIDs)==0")
		return nil
	}
}
func (ReminderDalGae) SetReminderStatus(ctx context.Context, reminderID, returnTransferID string, status string, when time.Time) (reminder models4debtus.Reminder, err error) {
	var (
		changed        bool
		previousStatus string
	)
	err = facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		if reminder, err = dtdal.Reminder.GetReminderByID(ctx, tx, reminderID); err != nil {
			return
		} else {
			switch status {
			case string(models4debtus.ReminderStatusDiscarded):
				reminder.Data.DtDiscarded = when
			case string(models4debtus.ReminderStatusSent):
				reminder.Data.DtSent = when
			case string(models4debtus.ReminderStatusSending):
				// pass
			case string(models4debtus.ReminderStatusViewed):
				reminder.Data.DtViewed = when
			case string(models4debtus.ReminderStatusUsed):
				reminder.Data.DtUsed = when
			default:
				return errors.New("unsupported status: " + status)
			}
			previousStatus = reminder.Data.Status
			changed = previousStatus != status
			if returnTransferID != "" && status == string(models4debtus.ReminderStatusDiscarded) {
				for _, id := range reminder.Data.ClosedByTransferIDs { // TODO: WTF are we doing here?
					if id == returnTransferID {
						logus.Infof(ctx, "new status: '%v', Reminder{Status: '%v', ClosedByTransferIDs: %v}", status, reminder.Data.Status, reminder.Data.ClosedByTransferIDs)
						return ErrDuplicateAttemptToDiscardReminder
					}
				}
				reminder.Data.ClosedByTransferIDs = append(reminder.Data.ClosedByTransferIDs, returnTransferID)
				changed = true
			}
			if changed {
				reminder.Data.Status = status
				if err = tx.Set(ctx, reminder.Record); err != nil {
					err = fmt.Errorf("failed to save reminder to db (id=%v): %w", reminderID, err)
				}
			}
			return
		}
	}, nil)
	if err == nil {
		if changed {
			logus.Debugf(ctx, "Reminder(id=%v) status changed from '%v' to '%v'", reminderID, previousStatus, status)
		} else {
			logus.Debugf(ctx, "Reminder(id=%v) status not changed as already '%v'", reminderID, status)
		}
	}
	return
}

var ErrDuplicateAttemptToDiscardReminder = errors.New("duplicate attempt to close reminder by same return transfer")
