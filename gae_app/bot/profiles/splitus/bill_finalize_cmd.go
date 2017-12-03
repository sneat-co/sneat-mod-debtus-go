package splitus

import (
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

const FINALIZE_BILL_COMMAND = "finalize_bill"

var finalizeBillCommand = billCallbackCommand(FINALIZE_BILL_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		footer := "<b>Are you ready to split the bill?</b>" +
			"\n" + "You won't be able to add/remove participants or change total once the bill is finalized."
		if m.Text, err = getBillCardMessageText(whc.Context(), whc.GetBotCode(), whc, bill, true, footer); err != nil {
			return
		}
		m.Format = bots.MessageFormatHTML
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         emoji.GREEN_CHECKBOX + " Yes, split the bill!",
					CallbackData: billCallbackCommandData(FINALIZE_BILL_COMMAND, bill.ID),
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
