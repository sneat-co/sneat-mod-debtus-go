package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/log"
	"net/url"
)

const CHANGE_BILL_TOTAL_COMMAND = "bill_total"

var changeBillTotalCommand = billCallbackCommand(CHANGE_BILL_TOTAL_COMMAND,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
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
