package splitus

import (
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/log"
)

const CLOSE_BILL_COMMAND = "close-bill"

var closeBillCommand = billCallbackCommand(CLOSE_BILL_COMMAND, nil,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "closeBillCommand.CallbackAction()")
		return ShowBillCard(whc, true, bill, "Sorry, not implemented yet.")
	},
)
