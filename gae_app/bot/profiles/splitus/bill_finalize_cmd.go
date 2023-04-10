package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"net/url"
)

const finalizeBillCommandCode = "finalize_bill"

var finalizeBillCommand = billCallbackCommand(finalizeBillCommandCode, nil,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		footer := "<b>Are you ready to split the bill?</b>" +
			"\n" + "You won't be able to add/remove participants or change total once the bill is finalized."
		if m.Text, err = getBillCardMessageText(whc.Context(), whc.GetBotCode(), whc, bill, true, footer); err != nil {
			return
		}
		m.Format = botsfw.MessageFormatHTML
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         emoji.GREEN_CHECKBOX + " Yes, split the bill!",
					CallbackData: billCallbackCommandData(finalizeBillCommandCode, bill.ID),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         emoji.NO_ENTRY_SIGN_ICON + " " + "Cancel",
					CallbackData: billCardCallbackCommandData(bill.ID),
				},
			},
		)
		return
	},
)
