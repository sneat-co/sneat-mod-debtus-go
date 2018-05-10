package admin

import (
	"testing"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
)

func TestSendFeedbackToAdmins(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()
	SendFeedbackToAdmins(nil, "", models.Feedback{})
}
