package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"golang.org/x/net/context"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
	"net/url"
	"strconv"
	"time"
)

func _validateSetReminderIsSentMessageIDs(messageIntID int64, messageStrID string, sentAt time.Time) error {
	if messageIntID != 0 && messageStrID != "" {
		return errors.New("messageIntID != 0 && messageStrID != ''")
	} else if messageIntID == 0 && messageStrID == "" {
		return errors.New("messageIntID == 0 && messageStrID == ''")
	}
	if sentAt.IsZero() {
		return errors.New("sentAt.IsZero()")
	}
	return nil
}

func (_ ReminderDalGae) DelaySetReminderIsSent(c context.Context, reminderID int64, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error {
	if err := _validateSetReminderIsSentMessageIDs(messageIntID, messageStrID, sentAt); err != nil {
		return err
	}
	if err := gae.CallDelayFunc(c, common.QUEUE_REMINDERS, "set-reminder-is-sent", delayedSetReminderIsSent, reminderID, sentAt, messageIntID, messageStrID, locale, errDetails); err != nil {
		return errors.Wrap(err, "Failed to delay execution of setReminderIsSent")
	}
	return nil
}

var delayedSetReminderIsSent = delay.Func("setReminderIsSent", setReminderIsSent)

func setReminderIsSent(c context.Context, reminderID int64, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error {
	return dal.Reminder.SetReminderIsSent(c, reminderID, sentAt, messageIntID, messageStrID, locale, errDetails)
}

func CreateSendReminderTask(c context.Context, reminderID int64) *taskqueue.Task {
	if reminderID == 0 {
		panic("reminderID == 0")
	}
	t := taskqueue.NewPOSTTask("/task-queue/send-reminder", url.Values{"id": []string{strconv.FormatInt(reminderID, 10)}})
	return t
}

func QueueSendReminder(c context.Context, reminderID int64, dueIn time.Duration) error {
	if dueIn < 3*time.Hour {
		task := CreateSendReminderTask(c, reminderID)
		if dueIn > time.Duration(0) {
			task.Delay = dueIn + (3 * time.Second)
		}
		if _, err := gae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			return errors.Wrapf(err, "Failed to add task(name='%v', delay=%v) to '%v' queue", task.Name, task.Delay, common.QUEUE_REMINDERS)
		}
	}
	return nil
}
