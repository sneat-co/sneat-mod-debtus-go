package dtb_transfer

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

func reportReminderIsActed(whc bots.WebhookContext, action string) {
	ga := whc.GA()
	if err := ga.Queue(ga.GaEvent(
		"reminders",
		action,
	)); err != nil {
		log.Errorf(whc.Context(), err.Error())
		err = nil
	}
}
