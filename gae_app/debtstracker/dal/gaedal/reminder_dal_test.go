package gaedal

import (
	"golang.org/x/net/context"
	"testing"
)

func TestNewReminderKey(t *testing.T) {
	const reminderID = 135
	testDatastoreIntKey(t, reminderID, NewReminderKey(context.Background(), reminderID))
}

func TestNewReminderIncompleteKey(t *testing.T) {
	testDatastoreIncompleteKey(t, NewReminderIncompleteKey(context.Background()))
}
