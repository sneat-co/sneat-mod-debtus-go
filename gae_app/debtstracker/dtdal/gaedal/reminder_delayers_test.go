package gaedal

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"context"
	"github.com/strongo/app"
)

func Test__validateSetReminderIsSentMessageIDs(t *testing.T) {
	var err error
	now := time.Now()
	if err = _validateSetReminderIsSentMessageIDs(0, "", now); err == nil {
		t.Error("Should fail: _validateSetReminderIsSentMessageIDs(0, '')")
	}
	if err = _validateSetReminderIsSentMessageIDs(1, "not empty", now); err == nil {
		t.Error("Should fail: _validateSetReminderIsSentMessageIDs(1, 'not empty')")
	}
	if err = _validateSetReminderIsSentMessageIDs(1, "", time.Time{}); err == nil {
		t.Error("Should fail as sentAt is zero")
		if !strings.Contains(err.Error(), "sentAt.IsZero()") {
			t.Error("Error message does not contain 'sentAt.IsZero()'")
		}
	}
}

func TestDelaySetReminderIsSent(t *testing.T) {
	var err error

	reminderDal := NewReminderDalGae()

	if err = reminderDal.DelaySetReminderIsSent(context.TODO(), 0, time.Now(), 1, "", strongo.LocaleCodeEnUS, ""); err == nil {
		t.Error("Should fail as reminder is 0")
	}
	if err = reminderDal.DelaySetReminderIsSent(context.TODO(), 1, time.Now(), 0, "", strongo.LocaleCodeEnUS, ""); err == nil {
		t.Error("Should fail as no message id supplied")
	}
	if err = reminderDal.DelaySetReminderIsSent(context.TODO(), 1, time.Now(), 1, "not empty", strongo.LocaleCodeEnUS, ""); err == nil {
		t.Error("Should fail as both int and string message ids supplied")
	}
	if err = reminderDal.DelaySetReminderIsSent(context.TODO(), 1, time.Time{}, 1, "not empty", strongo.LocaleCodeEnUS, ""); err == nil {
		t.Error("Should fail as both int and string message ids supplied")
	}
	if err = reminderDal.DelaySetReminderIsSent(context.TODO(), 1, time.Time{}, 1, "", strongo.LocaleCodeEnUS, ""); err == nil {
		t.Error("Should fail as both sentAt is zero")
	}

	//countOfCallsToDelay := 0
	//apphostgae.CallDelayFunc = func(c context.Context, queueName, subPath string, f *delay.Function, args ...interface{}) error {
	//	countOfCallsToDelay += 1
	//	return nil
	//}
	if err = reminderDal.DelaySetReminderIsSent(context.TODO(), 1, time.Now(), 1, "", strongo.LocaleCodeEnUS, ""); err != nil {
		t.Error(fmt.Errorf("should NOT fail: %w", err).Error())
	}
	//if countOfCallsToDelay != 1 {
	//	t.Errorf("Expeted to get 1 call to delay, got: %v", countOfCallsToDelay)
	//}
}
