package admin

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"testing"
)

func TestSendFeedbackToAdmins(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()
	SendFeedbackToAdmins(nil, "", models.Feedback{})
}
