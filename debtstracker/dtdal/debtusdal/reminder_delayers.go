package debtusdal

import (
	"context"
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-core-modules/core/queues"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal/delayer4debtus"
	"github.com/strongo/delaying"
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

func (ReminderDalGae) DelaySetReminderIsSent(ctx context.Context, reminderID string, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error {
	if reminderID == "" {
		return errors.New("reminderID == 0")
	}
	if sentAt.IsZero() {
		return errors.New("sentAt.IsZero()")
	}
	if err := _validateSetReminderIsSentMessageIDs(messageIntID, messageStrID, sentAt); err != nil {
		return err
	}
	if err := delayer4debtus.SetReminderIsSent.EnqueueWork(ctx, delaying.With(queues.QueueReminders, "set-reminder-is-sent", 0), reminderID, sentAt, messageIntID, messageStrID, locale, errDetails); err != nil {
		return fmt.Errorf("failed to delay execution of delayedSetReminderIsSent: %w", err)
	}
	return nil
}

func delayedSetReminderIsSent(ctx context.Context, reminderID string, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error {
	return dtdal.Reminder.SetReminderIsSent(ctx, reminderID, sentAt, messageIntID, messageStrID, locale, errDetails)
}

func CreateSendReminderTask(_ context.Context, reminderID string) (err error) {
	return errors.New("TODO: implement CreateSendReminderTask")
	//if reminderID == "" {
	//	panic("reminderID == 0")
	//}
	//t := taskqueue.NewPOSTTask("/task-queue/send-reminder", url.Values{"id": []string{reminderID}})
	//return t
}

func QueueSendReminder(ctx context.Context, reminderID string, dueIn time.Duration) error {
	if dueIn < 3*time.Hour {
		/*task,*/ err := CreateSendReminderTask(ctx, reminderID)
		return err
		//if dueIn > time.Duration(0) {
		//	task.Delay = dueIn + (3 * time.Second)
		//}
		//if _, err := apphostgae.AddTaskToQueue(ctx, task, queues.QueueReminders); err != nil {
		//	return fmt.Errorf("failed to add task(name='%v', delay=%v) to '%v' queue: %w", task.Name, task.Delay, queues.QueueReminders, err)
		//}
	}
	return nil
}
