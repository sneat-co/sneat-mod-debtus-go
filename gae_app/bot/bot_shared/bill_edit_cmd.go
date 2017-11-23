package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/core"
	"net/url"
)

const EDIT_BILL_COMMAND = "edit_bill"

func EditBillCommand(params BotParams) bots.Command {
	return BillCallbackCommand(EDIT_BILL_COMMAND,
		func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
			c := whc.Context()
			log.Debugf(c, "editBillCommand.CallbackAction()")
			var mt string

			if mt, err = GetBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, ""); err != nil {
				return
			}
			if m, err = whc.NewEditMessage(mt, bots.MessageFormatHTML); err != nil {
				return
			}
			m.Keyboard = params.GetPrivateBillCardInlineKeyboard(whc, whc.GetBotCode(), bill)
			return
		},
	)
}
