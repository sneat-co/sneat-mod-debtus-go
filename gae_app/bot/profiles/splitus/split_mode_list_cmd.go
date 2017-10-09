package splitus

import (
	"fmt"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
)

var billSplitModesListCommand = bot_shared.BillCallbackCommand("split-modes",
	func(whc bots.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "billSplitModesListCommand.CallbackAction()")
		var mt string
		if mt, err = bot_shared.GetBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, ""); err != nil {
			return
		}
		if m, err = whc.NewEditMessage(mt, bots.MessageFormatHTML); err != nil {
			return
		}
		callbackData := fmt.Sprintf("split-mode?bill=%v&mode=", bill.ID)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.SPLIT_MODE_EQUALLY),
					CallbackData: callbackData + string(models.SplitModeEqually),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text: whc.Translate(trans.SPLIT_MODE_PERCENTAGE),
					CallbackData: callbackData + string(models.SplitModePercentage),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text: whc.Translate(trans.SPLIT_MODE_SHARES),
					CallbackData: callbackData + string(models.SplitModeShare),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text: whc.Translate(trans.SPLIT_MODE_EXACT_AMOUNT),
					CallbackData: callbackData + string(models.SplitModeExactAmount),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.BUTTON_TEXT_CANCEL),
					CallbackData: bot_shared.BillCardCallbackCommandData(bill.ID),
				},
			},
		)
		m.Keyboard = keyboard
		return
	},
)
