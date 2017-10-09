package splitus

import (
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/DebtsTracker/translations/emoji"
	"fmt"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
)

func getGroupBillCardInlineKeyboard(translator strongo.SingleLocaleTranslator, bill models.Bill) *tgbotapi.InlineKeyboardMarkup {
	//	//{{Text: "I paid for the bill alone", CallbackData: joinBillCallbackPrefix + "&i=paid-alone"}},
	//	//{{Text:"I paid part of this bill",CallbackData:  joinBillCallbackPrefix + "&i=paid-part"}},
	//	//{{Text: "I owe for this bill", CallbackData: joinBillCallbackPrefix + "&i=owe"}},
	//	//{{Text: "I don't share this bill", CallbackData: BillCallbackCommandData(LEAVE_BILL_COMMAND, bill.ID)}},
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         translator.Translate(trans.BUTTON_TEXT_MANAGE_MEMBERS),
					CallbackData: bot_shared.GetBillMembersCallbackData(bill.ID),
				},
			},
			{
				{
					Text:         translator.Translate(trans.BUTTON_TEXT_SPLIT_MODE, translator.Translate(string(bill.SplitMode))),
					CallbackData: bot_shared.BillCallbackCommandData(BILL_SPLIT_COMMAND, bill.ID),
				},
			},
			{
				{
					Text:         emoji.GREEN_CHECKBOX + " Finalize bill",
					CallbackData: bot_shared.BillCallbackCommandData(FINALIZE_BILL_COMMAND, bill.ID),
				},
			},
		},
	}
}

func getPrivateBillCardInlineKeyboard(translator strongo.SingleLocaleTranslator, botCode string, bill models.Bill) *tgbotapi.InlineKeyboardMarkup {
	callbackData := fmt.Sprintf("split-mode?bill=%v&mode=", bill.ID)
	return tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         translator.Translate(trans.BUTTON_TEXT_MANAGE_MEMBERS),
				CallbackData: bot_shared.GetBillMembersCallbackData(bill.ID),
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         translator.Translate(trans.BUTTON_TEXT_CHANGE_BILL_PAYER),
				CallbackData: fmt.Sprintf(CHANGE_BILL_PAYER_COMMAND+"?bill=%v", bill.ID)},
			{
				Text:         translator.Translate(trans.BUTTON_TEXT_SPLIT_MODE, translator.Translate(string(bill.SplitMode))),
				CallbackData: fmt.Sprintf("split-modes?bill=%v", bill.ID),
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         translator.Translate("💯 Change total"),
				CallbackData: bot_shared.BillCallbackCommandData(bot_shared.CHANGE_BILL_TOTAL_COMMAND, bill.ID),
			},
			{
				Text:         translator.Translate("✍ Adjust per person"),
				CallbackData: callbackData + string(models.SplitModePercentage),
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         translator.Translate("📝 Комментарий"),
				CallbackData: bot_shared.BillCallbackCommandData(bot_shared.ADD_BILL_COMMENT_COMMAND, bill.ID),
			},
			{
				Text:         translator.Translate(trans.BUTTON_TEXT_FINALIZE_BILL),
				CallbackData: bot_shared.BillCallbackCommandData(bot_shared.CLOSE_BILL_COMMAND, bill.ID),
			},
		},
	)
}
