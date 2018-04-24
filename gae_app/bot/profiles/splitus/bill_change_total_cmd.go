package splitus

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

const CHANGE_BILL_TOTAL_COMMAND = "bill_total"

var changeBillTotalCommand = billCallbackCommand(CHANGE_BILL_TOTAL_COMMAND, nil,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "changeBillTotalCommand.CallbackAction()")
		//var editedMessage *tgbotapi.EditMessageTextConfig
		//if editedMessage, err = dtb_common.NewTelegramEditMessage(whc, "Enter new total for the bill:"); err != nil {
		//	return
		//}
		//
		//editedMessage.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
		m = whc.NewMessage("Enter new total for the bill:")
		m.Keyboard = &tgbotapi.ForceReply{ForceReply: true, Selective: true}
		return
	},
)
