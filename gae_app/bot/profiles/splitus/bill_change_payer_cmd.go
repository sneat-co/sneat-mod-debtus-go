package splitus

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app/log"
	"bitbucket.com/debtstracker/gae_app/bot/bot_shared"
)

const CHANGE_BILL_PAYER_COMMAND = "change-bill-payer"

var changeBillPayerCommand = bot_shared.BillCallbackCommand(CHANGE_BILL_PAYER_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "changeBillPayerCommand.CallbackAction()")
		var (
			mt            string
			//editedMessage *tgbotapi.EditMessageTextConfig
		)
		if mt, err = bot_shared.GetBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, whc.Translate(trans.MESSAGE_TEXT_BILL_ASK_WHO_PAID)); err != nil {
			return
		}
		if m, err = whc.NewEditMessage(mt, bots.MessageFormatHTML); err != nil {
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
					CallbackData: bot_shared.BillCardCallbackCommandData(bill.ID),
				},
			})
		}

		markup.InlineKeyboard = append(markup.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BUTTON_TEXT_CANCEL),
				CallbackData: bot_shared.BillCardCallbackCommandData(bill.ID),
			},
		})

		m.Keyboard = markup
		return
	},
)
