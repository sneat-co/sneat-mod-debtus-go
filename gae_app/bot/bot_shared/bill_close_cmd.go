package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/app/log"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
)

const CLOSE_BILL_COMMAND = "close-bill"

func CloseBillCommand(botParams BotParams) bots.Command {
	return BillCallbackCommand(CLOSE_BILL_COMMAND,
		func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			log.Debugf(c, "closeBillCommand.CallbackAction()")
			return ShowBillCard(whc, botParams, true, bill, "Sorry, not implemented yet.")
		},
	)
}