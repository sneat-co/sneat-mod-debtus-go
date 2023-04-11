package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/log"
	"net/url"
)

var billSplitModesListCommand = billCallbackCommand("split-modes",
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, bill models.Bill) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "billSplitModesListCommand.CallbackAction()")
		var mt string
		if mt, err = getBillCardMessageText(c, whc.GetBotCode(), whc, bill, true, ""); err != nil {
			return
		}
		if m, err = whc.NewEditMessage(mt, botsfw.MessageFormatHTML); err != nil {
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
					Text:         whc.Translate(trans.SPLIT_MODE_PERCENTAGE),
					CallbackData: callbackData + string(models.SplitModePercentage),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.SPLIT_MODE_SHARES),
					CallbackData: callbackData + string(models.SplitModeShare),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.SPLIT_MODE_EXACT_AMOUNT),
					CallbackData: callbackData + string(models.SplitModeExactAmount),
				},
			},
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         whc.Translate(trans.BUTTON_TEXT_CANCEL),
					CallbackData: billCardCallbackCommandData(bill.ID),
				},
			},
		)
		m.Keyboard = keyboard
		return
	},
)
