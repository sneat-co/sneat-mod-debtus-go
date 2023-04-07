package dtb_transfer

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_inline"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

func showReceiptAnnouncement(whc bots.WebhookContext, receiptID int64, creatorName string) (m bots.MessageFromBot, err error) {
	var inlineMessageID string
	input := whc.Input()
	switch input.(type) {
	case bots.WebhookChosenInlineResult:
		inlineMessageID = input.(bots.WebhookChosenInlineResult).GetInlineMessageID()
	case bots.WebhookCallbackQuery:
		inlineMessageID = input.(bots.WebhookCallbackQuery).GetInlineMessageID()
	default:
		return m, fmt.Errorf("showReceiptAnnouncement: Unsupported InputType=%T", input)
	}

	c := whc.Context()

	receipt, err := dtdal.Receipt.GetReceiptByID(c, receiptID)
	if err != nil {
		return m, err
	}
	if creatorName == "" {
		user, err := facade.User.GetUserByID(c, receipt.CreatorUserID)
		if err != nil {
			return m, err
		}
		creatorName = user.FullName()
	}

	messageText := getInlineReceiptMessageText(whc, whc.GetBotCode(), whc.Locale().Code5, creatorName, receiptID)
	m, err = whc.NewEditMessage(messageText, bots.MessageFormatHTML)
	m.EditMessageUID = telegram.NewInlineMessageUID(inlineMessageID)
	m.DisableWebPagePreview = true
	kbRows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData(
				whc.Translate(trans.COMMAND_TEXT_VIEW_RECEIPT_DETAILS),
				fmt.Sprintf("%v?id=%v&locale=%v",
					VIEW_RECEIPT_IN_TELEGRAM_COMMAND, common.EncodeID(receiptID), whc.Locale().Code5,
				),
			),
		},
	}
	kbRows = append(kbRows, dtb_inline.GetChooseLangInlineKeyboard(
		fmt.Sprintf("%v?id=%v", CHANGE_RECEIPT_LANG_COMMAND, common.EncodeID(receiptID))+"&locale=%v",
		whc.Locale().Code5,
	)...)
	m.Keyboard = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: kbRows,
	}
	return
}

const VIEW_RECEIPT_IN_TELEGRAM_COMMAND = "tg-view-receipt"

func GetUrlForReceiptInTelegram(botCode string, receiptID int64, localeCode5 string) string {
	return fmt.Sprintf("https://t.me/%v?start=receipt-%v-view_%v", botCode, receiptID, localeCode5)
}
