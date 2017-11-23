package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/core"
	"net/url"
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
