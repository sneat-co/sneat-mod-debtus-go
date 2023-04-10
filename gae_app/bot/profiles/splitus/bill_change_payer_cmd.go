package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/log"
	"net/url"
)

const CHANGE_BILL_PAYER_COMMAND = "change-bill-payer"

var changeBillPayerCommand = billCallbackCommand(CHANGE_BILL_PAYER_COMMAND, nil,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "changeBillPayerCommand.CallbackAction()")
		var (
			mt string
			//editedMessage *tgbotapi.EditMessageTextConfig
		)
		if mt, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, whc.Translate(trans.MESSAGE_TEXT_BILL_ASK_WHO_PAID)); err != nil {
			return
		}
		if m, err = whc.NewEditMessage(mt, botsfw.MessageFormatHTML); err != nil {
			return
		}
		markup := tgbotapi.NewInlineKeyboardMarkup()

		for _, member := range bill.GetBillMembers() {
			s := member.Name
			if member.Paid > 0 {
				s = "✔ " + s
			}

			markup.InlineKeyboard = append(markup.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
				{
					Text:         s,
					CallbackData: billCardCallbackCommandData(bill.ID),
				},
			})
		}

		markup.InlineKeyboard = append(markup.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BUTTON_TEXT_CANCEL),
				CallbackData: billCardCallbackCommandData(bill.ID),
			},
		})

		m.Keyboard = markup
		return
	},
)
