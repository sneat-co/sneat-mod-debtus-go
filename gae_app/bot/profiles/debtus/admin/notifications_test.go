package admin

import (
	"context"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
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
