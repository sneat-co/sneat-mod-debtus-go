package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/log"
	"net/url"
)

const editBillCommandCode = "edit_bill"

var editBillCommand = billCallbackCommand(editBillCommandCode,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "editBillCommand.CallbackAction()")
		var mt string

		if mt, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, ""); err != nil {
			return
		}
		if m, err = whc.NewEditMessage(mt, botsfw.MessageFormatHTML); err != nil {
			return
		}
		m.Keyboard = getPrivateBillCardInlineKeyboard(whc, whc.GetBotCode(), bill)
		return
	},
)
