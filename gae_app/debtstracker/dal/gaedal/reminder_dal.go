package gaedal

import (
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/datastore"
)

func NewReminderIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.ReminderKind, nil)
}

func NewReminderKey(c context.Context, reminderID int64) *datastore.Key {
	if reminderID == 0 {
		panic("reminderID == 0")
	}
	return gaedb.NewKey(c, models.ReminderKind, "", reminderID, nil)
}

type ReminderDalGae struct {
}

func NewReminderDalGae() ReminderDalGae {
	return ReminderDalGae{}
}

var _ dal.ReminderDal = (*ReminderDalGae)(nil)

func (reminderDalGae ReminderDalGae) GetReminderByID(c context.Context, id int64) (models.Reminder, error) {
	var reminderEntity models.ReminderEntity
	err := gaedb.Get(c, NewReminderKey(c, id), &reminderEntity)
	if err == datastore.ErrNoSuchEntity {
		err = db.NewErrNotFoundByIntID(models.ReminderKind, id, err)
	} else if err != nil {
		err = errors.Wrapf(err, "Failed to get reminder by id=%v", id)
	}
	return models.NewReminder(id, &reminderEntity), err
}

func (reminderDalGae ReminderDalGae) SaveReminder(c context.Context, reminder models.Reminder) (err error) {
	_, err = gaedb.Put(c, NewReminderKey(c, reminder.ID), reminder.ReminderEntity)
	return
}

func (reminderDalGae ReminderDalGae) GetSentReminderIDsByTransferID(c context.Context, transferID int64) ([]int64, error) {
	q := datastore.NewQuery(models.ReminderKind)
	q = q.Filter("TransferID =", transferID)
	q = q.Filter("Status =", models.ReminderStatusSent)
	q = q.KeysOnly()
	if keys, err := q.GetAll(c, nil); err != nil {
		return nil, errors.Wrapf(err, "Failed to get sent reminders by transfer id=%v", transferID)
	} else {
		reminderIDs := make([]int64, len(keys))
		for i, key := range keys {
			reminderIDs[i] = key.IntID()
		}
		return reminderIDs, nil
	}
}

func (reminderDalGae ReminderDalGae) GetActiveReminderIDsByTransferID(c context.Context, transferID int64) ([]int64, error) {
	q := datastore.NewQuery(models.ReminderKind)
	q = q.Filter("TransferID =", transferID)
	q = q.Filter("DtNext >", time.Time{})
	q = q.KeysOnly()
	if keys, err := q.GetAll(c, nil); err != nil {
		return nil, errors.Wrapf(err, "Failed to get active reminders by transfer id=%v", transferID)
	} else {
		reminderIDs := make([]int64, len(keys))
		for i, key := range keys {
			reminderIDs[i] = key.IntID()
		}
		return reminderIDs, nil
	}
}

func (reminderDalGae ReminderDalGae) SetReminderIsSent(c context.Context, reminderID int64, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error {
	gaehost.GaeLogger.Debugf(c, "setReminderIsSent(reminderID=%v, sentAt=%v, messageIntID=%v, messageStrID=%v)", reminderID, sentAt, messageIntID, messageStrID)
	if err := _validateSetReminderIsSentMessageIDs(messageIntID, messageStrID, sentAt); err != nil {
		return err
	}
	reminder := models.NewReminder(reminderID, nil)
	return dal.DB.RunInTransaction(c, func(c context.Context) error {
		return reminderDalGae.SetReminderIsSentInTransaction(c, reminder, sentAt, messageIntID, messageStrID, locale, errDetails)
	}, nil)
}

func (reminderDalGae ReminderDalGae) SetReminderIsSentInTransaction(c context.Context, reminder models.Reminder, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) (err error) {
	if reminder.ReminderEntity == nil {
		reminder, err = reminderDalGae.GetReminderByID(c, reminder.ID)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				return nil
			}
			return errors.Wrap(err, "Failed to get reminder by ID")
		}
	}
	if reminder.Status != models.ReminderStatusSending {
		log.Errorf(c, "reminder.Status:%v != models.ReminderStatusSending:%v", reminder.Status, models.ReminderStatusSending)
		return nil
	} else {
		reminder.Status = models.ReminderStatusSent
		reminder.DtSent = sentAt
		reminder.DtScheduled = reminder.DtNext
		reminder.DtNext = time.Time{}
		reminder.ErrDetails = errDetails
		reminder.Locale = locale
		if messageIntID != 0 {
			reminder.MessageIntID = messageIntID
		}
		if messageStrID != "" {
			reminder.MessageStrID = messageStrID
		}
		if _, err = gaedb.Put(c, NewReminderKey(c, reminder.ID), reminder.ReminderEntity); err != nil {
			err = errors.Wrap(err, "Failed to save reminder to datastore")
		}
		return err
	}
}

func (reminderDalGae ReminderDalGae) RescheduleReminder(c context.Context, reminderID int64, remindInDuration time.Duration) (oldReminder, newReminder models.Reminder, err error) {
	var (
		newReminderKey    *datastore.Key
		newReminderEntity *models.ReminderEntity
	)
	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		oldReminder, err = reminderDalGae.GetReminderByID(c, reminderID)
		if err != nil {
			return errors.Wrap(err, "Failed to get oldReminder by id")
		}
		if oldReminder.IsRescheduled {
			err = dal.ErrReminderAlreadyRescheduled
			return err
		}
		reminderKey := NewReceiptKey(c, reminderID)
		if remindInDuration == time.Duration(0) {
			if _, err = gaedb.Put(tc, reminderKey, oldReminder.ReminderEntity); err != nil {
				return err
			}
		} else {
			nextReminderOn := time.Now().Add(remindInDuration)
			newReminderEntity = oldReminder.ScheduleNextReminder(reminderID, nextReminderOn)
			newReminderKey = NewReminderIncompleteKey(tc)
			keys, err := gaedb.PutMulti(tc, []*datastore.Key{reminderKey, newReminderKey}, []interface{}{oldReminder.ReminderEntity, newReminderEntity})
			if err != nil {
				err = errors.Wrap(err, "Failed to reschedule oldReminder")
			}
			newReminderKey = keys[1]
			if err = QueueSendReminder(tc, newReminderKey.IntID(), remindInDuration); err != nil { // TODO: Should be outside of DAL?
				return err
			}
		}
		return err
	}, dal.CrossGroupTransaction)
	if err != nil {
		return
	}
	if newReminderKey != nil && newReminderEntity != nil {
		newReminder = models.Reminder{
			IntegerID:      db.NewIntID(newReminderKey.IntID()),
			ReminderEntity: newReminderEntity,
		}
	}
	return
}
