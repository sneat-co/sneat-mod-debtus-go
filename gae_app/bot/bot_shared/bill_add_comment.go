package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/app/log"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
)

const ADD_BILL_COMMENT_COMMAND = "bill_comment"

var addBillComment = BillCallbackCommand(ADD_BILL_COMMENT_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "addBillComment.CallbackAction()")

		//var editedMessage *tgbotapi.EditMessageTextConfig
		//if editedMessage, err = dtb_common.NewTelegramEditMessage(whc, "Enter new total for the bill:"); err != nil {
		//	return
		//}
		//
		//editedMessage.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: true}
		m = whc.NewMessage("Send your comment:")
		m.Keyboard = &tgbotapi.ForceReply{ForceReply: true, Selective: true}
		return
	},
)
