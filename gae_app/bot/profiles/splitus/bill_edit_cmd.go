package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"net/url"
)

const editBillCommandCode = "edit_bill"

var editBillCommand = billCallbackCommand(editBillCommandCode, nil,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "editBillCommand.CallbackAction()")
		var mt string

		if mt, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, ""); err != nil {
			return
		}
		if m, err = whc.NewEditMessage(mt, bots.MessageFormatHTML); err != nil {
			return
		}
		m.Keyboard = getPrivateBillCardInlineKeyboard(whc, whc.GetBotCode(), bill)
		return
	},
)
