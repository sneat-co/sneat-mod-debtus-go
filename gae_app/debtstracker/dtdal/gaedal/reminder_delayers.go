package gaedal

import (
	"fmt"
	apphostgae "github.com/strongo/app-host-gae"
	"net/url"
	"strconv"
	"time"

	"context"
	"errors"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
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

func (ReminderDalGae) DelaySetReminderIsSent(c context.Context, reminderID int, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error {
	if err := _validateSetReminderIsSentMessageIDs(messageIntID, messageStrID, sentAt); err != nil {
		return err
	}
	if err := apphostgae.CallDelayFunc(c, common.QUEUE_REMINDERS, "set-reminder-is-sent", delayedSetReminderIsSent, reminderID, sentAt, messageIntID, messageStrID, locale, errDetails); err != nil {
		return fmt.Errorf("failed to delay execution of setReminderIsSent: %w", err)
	}
	return nil
}

var delayedSetReminderIsSent = delay.Func("setReminderIsSent", setReminderIsSent)

func setReminderIsSent(c context.Context, reminderID int, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error {
	return dtdal.Reminder.SetReminderIsSent(c, reminderID, sentAt, messageIntID, messageStrID, locale, errDetails)
}

func CreateSendReminderTask(c context.Context, reminderID int) *taskqueue.Task {
	if reminderID == 0 {
		panic("reminderID == 0")
	}
	t := taskqueue.NewPOSTTask("/task-queue/send-reminder", url.Values{"id": []string{strconv.Itoa(reminderID)}})
	return t
}

func QueueSendReminder(c context.Context, reminderID int, dueIn time.Duration) error {
	if dueIn < 3*time.Hour {
		task := CreateSendReminderTask(c, reminderID)
		if dueIn > time.Duration(0) {
			task.Delay = dueIn + (3 * time.Second)
		}
		if _, err := apphostgae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			return fmt.Errorf("failed to add task(name='%v', delay=%v) to '%v' queue: %w", task.Name, task.Delay, common.QUEUE_REMINDERS, err)
		}
	}
	return nil
}
