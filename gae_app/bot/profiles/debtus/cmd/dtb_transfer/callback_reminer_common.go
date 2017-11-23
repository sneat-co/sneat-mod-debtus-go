package dtb_transfer

import (
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/measurement-protocol"
)

func reportReminderIsActed(whc bots.WebhookContext, action string) {
	if err := whc.GaMeasurement().Queue(measurement.NewEvent(
		"reminders",
		action,
		whc.GaCommon(),
	)); err != nil {
		log.Errorf(whc.Context(), err.Error())
		err = nil
	}
}
