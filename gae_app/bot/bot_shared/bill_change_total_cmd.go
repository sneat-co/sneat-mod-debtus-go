package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
)

const CHANGE_BILL_TOTAL_COMMAND = "bill_total"

var changeBillTotalCommand = BillCallbackCommand(CHANGE_BILL_TOTAL_COMMAND,
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