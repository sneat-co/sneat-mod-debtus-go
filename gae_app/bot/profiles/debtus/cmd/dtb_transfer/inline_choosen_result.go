package dtb_transfer

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_inline"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/trans"
)

func showReceiptAnnouncement(whc botsfw.WebhookContext, receiptID int, creatorName string) (m botsfw.MessageFromBot, err error) {
	var inlineMessageID string
	switch input := whc.Input().(type) {
	case botsfw.WebhookChosenInlineResult:
		inlineMessageID = input.GetInlineMessageID()
	case botsfw.WebhookCallbackQuery:
		inlineMessageID = input.GetInlineMessageID()
	default:
		return m, fmt.Errorf("showReceiptAnnouncement: Unsupported InputType=%T", input)
	}

	c := whc.Context()

	receipt, err := dtdal.Receipt.GetReceiptByID(c, nil, receiptID)
	if err != nil {
		return m, err
	}
	if creatorName == "" {
		user, err := facade.User.GetUserByID(c, nil, receipt.Data.CreatorUserID)
		if err != nil {
			return m, err
		}
		creatorName = user.Data.FullName()
	}

	messageText := getInlineReceiptMessageText(whc, whc.GetBotCode(), whc.Locale().Code5, creatorName, receiptID)
	m, err = whc.NewEditMessage(messageText, botsfw.MessageFormatHTML)
	m.EditMessageUID = telegram.NewInlineMessageUID(inlineMessageID)
	m.DisableWebPagePreview = true
	kbRows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData(
				whc.Translate(trans.COMMAND_TEXT_VIEW_RECEIPT_DETAILS),
				fmt.Sprintf("%v?id=%v&locale=%v",
					VIEW_RECEIPT_IN_TELEGRAM_COMMAND, common.EncodeIntID(receiptID), whc.Locale().Code5,
				),
			),
		},
	}
	kbRows = append(kbRows, dtb_inline.GetChooseLangInlineKeyboard(
		fmt.Sprintf("%v?id=%v", CHANGE_RECEIPT_LANG_COMMAND, common.EncodeIntID(receiptID))+"&locale=%v",
		whc.Locale().Code5,
	)...)
	m.Keyboard = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: kbRows,
	}
	return
}

const VIEW_RECEIPT_IN_TELEGRAM_COMMAND = "tg-view-receipt"

func GetUrlForReceiptInTelegram(botCode string, receiptID int, localeCode5 string) string {
	return fmt.Sprintf("https://t.me/%v?start=receipt-%v-view_%v", botCode, receiptID, localeCode5)
}
