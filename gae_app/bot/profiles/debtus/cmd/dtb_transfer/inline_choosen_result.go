package dtb_transfer

import (
	"fmt"
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_inline"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
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

	receipt, err := dal.Receipt.GetReceiptByID(c, receiptID)
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

var ViewReceiptInTelegramCallbackCommand = bots.NewCallbackCommand(
	VIEW_RECEIPT_IN_TELEGRAM_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "ViewReceiptInTelegramCallbackCommand.CallbackAction()")
		query := callbackUrl.Query()
		receiptID, err := common.DecodeID(query.Get("id"))
		if err != nil {
			return m, err
		}
		receipt, err := dal.Receipt.GetReceiptByID(c, receiptID)
		if err != nil {
			return m, err
		}
		currentUserID := whc.AppUserIntID()
		if receipt.CreatorUserID != currentUserID {
			if receipt.CounterpartyUserID == 0 {
				linker := facade.NewReceiptUsersLinker(nil) // TODO: Link users
				if _, err = linker.LinkReceiptUsers(c, receiptID, currentUserID); err != nil {
					return m, err
				}
			} else if receipt.CounterpartyUserID != currentUserID {
				// TODO: Should we allow to see receipt but block from changing it?
				log.Warningf(c, `Security issue: receipt.CreatorUserID != currentUserID && receipt.CounterpartyUserID != currentUserID
	currentUserID: %d
	receipt.CreatorUserID: %d
	receipt.CounterpartyUserID: %d
				`, currentUserID, receipt.CreatorUserID, receipt.CounterpartyUserID)
			} else {
				// receipt.CounterpartyUserID == currentUserID - we are fine
			}
		}
		localeCode5 := query.Get("locale")
		if len(localeCode5) != 5 {
			return m, errors.New("len(localeCode5) != 5")
		}

		callbackAnswer := tgbotapi.NewCallbackWithURL(
			GetUrlForReceiptInTelegram(whc.GetBotCode(), receiptID, localeCode5),
			//common.GetReceiptUrlForUser(
			//	receiptID,
			//	whc.AppUserIntID(),
			//	whc.BotPlatform().ID(),
			//	whc.GetBotCode(),
			//) + "&lang=" + localeCode5,
		)
		m.BotMessage = telegram.CallbackAnswer(callbackAnswer)
		// TODO: https://core.telegram.org/bots/api#answercallbackquery, show_alert = true
		return
	},
)
