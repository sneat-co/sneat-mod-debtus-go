package admin

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"testing"
)

func TestSendFeedbackToAdmins(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()
	_ = SendFeedbackToAdmins(context.Background(), "", models.Feedback{})
}
