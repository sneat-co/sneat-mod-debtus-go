package dtb_transfer

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"github.com/strongo/gamp"
)

func reportReminderIsActed(whc bots.WebhookContext, action string) {
	if err := whc.GaMeasurement().Queue(gamp.NewEvent(
		"reminders",
		action,
		whc.GaCommon(),
	)); err != nil {
		log.Errorf(whc.Context(), err.Error())
		err = nil
	}
}
