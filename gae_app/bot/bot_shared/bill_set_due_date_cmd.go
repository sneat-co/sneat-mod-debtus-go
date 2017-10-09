package bot_shared

import (
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
)

const SET_BILL_DUE_DATE_COMMAND = "bill_due"

var setBillDueDateCommand = bots.Command{
	Code: SET_BILL_DUE_DATE_COMMAND,
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		chatEntity := whc.ChatEntity()
		chatEntity.SetAwaitingReplyTo(SET_BILL_DUE_DATE_COMMAND)
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
