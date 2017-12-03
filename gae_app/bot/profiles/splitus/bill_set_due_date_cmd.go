package splitus

import (
	"net/url"

	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

const setBillDueDateCommandCode = "bill_due"

var setBillDueDateCommand = bots.Command{
	Code: setBillDueDateCommandCode,
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		chatEntity := whc.ChatEntity()
		chatEntity.SetAwaitingReplyTo(setBillDueDateCommandCode)
		chatEntity.AddWizardParam("bill", callbackUrl.Query().Get("id"))
		log.Debugf(c, "setBillDueDateCommand.CallbackAction()")
		m = whc.NewMessage("Please set bill due date as dd.mm.yyyy")
		m.Keyboard = &tgbotapi.ForceReply{ForceReply: true, Selective: true}
		return
	},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "setBillDueDateCommand.Action()")
		m = whc.NewMessage("Not implemented yet")
		return
	},
}
