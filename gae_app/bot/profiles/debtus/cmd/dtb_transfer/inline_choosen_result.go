package dtb_transfer

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_inline"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	//"github.com/DebtsTracker/translations/emoji"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	//"bitbucket.com/asterus/debtstracker-server/gae_app/invites"
	//"fmt"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"net/url"
	"time"
)

func CreateReceiptAndEditMessage(whc bots.WebhookContext, inlineMessageID string, transferID int64, creatorName string) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	transfer, err := dal.Transfer.GetTransferByID(c, transferID)
	if err != nil {
		return m, err
	}
	receipt := models.NewReceiptEntity(whc.AppUserIntID(), transferID, transfer.Counterparty().UserID, whc.Locale().Code5, telegram_bot.TelegramPlatformID, "", general.CreatedOn{
		CreatedOnID:       whc.GetBotCode(), // TODO: Replace with method call.
		CreatedOnPlatform: whc.BotPlatform().Id(),
	})
	receiptID, err := dal.Receipt.CreateReceipt(c, &receipt)
	if err != nil {
		return m, err
	}

	dal.Receipt.DelayedMarkReceiptAsSent(c, receiptID, transferID, time.Now())
	return showReceiptAnnouncement(whc, receiptID, creatorName)
}

const CHANGE_RECEIPT_LANG_COMMAND = "change-lang-receipt"

var ChangeReceiptAnnouncementLangCommand = bots.NewCallbackCommand(
	CHANGE_RECEIPT_LANG_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		query := callbackUrl.Query()
		code5 := query.Get("locale")
		if len(code5) != 5 {
			return m, errors.New("ChangeReceiptAnnouncementLangCommand: len(code5) != 5")
		}
		whc.SetLocale(code5)
		receiptID, err := common.DecodeID(query.Get("id"))
		if err != nil {
			return m, err
		}
		c := whc.Context()
		receipt, err := dal.Receipt.GetReceiptByID(c, receiptID)
		if err != nil {
			return m, err
		}
		user, err := dal.User.GetUserByID(c, receipt.CreatorUserID)
		if err != nil {
			return m, err
		}
		return showReceiptAnnouncement(whc, receiptID, user.FullName())
	},
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
		return m, errors.New(fmt.Sprintf("showReceiptAnnouncement: Unsupported InputType=%T", input))
	}

	receiptUrl := getReceiptUrl(whc.GetBotCode(), common.EncodeID(receiptID), whc.Locale().Code5)

	messageText := getInlineReceiptAnnouncementMessage(whc, true, creatorName, receiptUrl)
	m, err = whc.NewEditMessage(messageText, bots.MessageFormatHTML)
	m.EditMessageUID = telegram_bot.NewInlineMessageUID(inlineMessageID)
	m.DisableWebPagePreview = true
	m.Keyboard = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData(
					whc.Translate(trans.COMMAND_TEXT_VIEW_RECEIPT_DETAILS),
					fmt.Sprintf("%v?id=%v&locale=%v",
						VIEW_RECEIPT_IN_TELEGRAM_COMMAND, common.EncodeID(receiptID), whc.Locale().Code5,
					),
				),
			},
			dtb_inline.GetChooseLangInlineKeyboard(
				fmt.Sprintf("%v?id=%v", CHANGE_RECEIPT_LANG_COMMAND, common.EncodeID(receiptID))+"&locale=%v",
				whc.Locale().Code5,
			),
		},
	}
	return
}

const VIEW_RECEIPT_IN_TELEGRAM_COMMAND = "tg-view-receipt"

func getReceiptUrl(botCode, receiptID, localeCode5 string) string {
	return fmt.Sprintf("https://t.me/%v?start=receipt-%v-view_%v", botCode, receiptID, localeCode5)
}

var ViewReceiptInTelegramCallbackCommand = bots.NewCallbackCommand(
	VIEW_RECEIPT_IN_TELEGRAM_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		query := callbackUrl.Query()
		receiptID, err := common.DecodeID(query.Get("id"))
		if err != nil {
			return m, err
		}
		c := whc.Context()
		receipt, err := dal.Receipt.GetReceiptByID(c, receiptID)
		if err != nil {
			return m, err
		}
		currentUserID := whc.AppUserIntID()
		if receipt.CreatorUserID != currentUserID {
			if receipt.CounterpartyUserID == 0 {
				linker := facade.ReceiptUsersLinker{} // TODO: Link users
				if err = linker.LinkReceiptUsers(c, receiptID, currentUserID); err != nil {
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

		callbackAnswer := tgbotapi.NewCallbackWithUrl(
			getReceiptUrl(whc.GetBotCode(), common.EncodeID(receiptID), localeCode5),
			//common.GetReceiptUrlForUser(
			//	receiptID,
			//	whc.AppUserIntID(),
			//	whc.BotPlatform().Id(),
			//	whc.GetBotCode(),
			//) + "&lang=" + localeCode5,
		)
		m.BotMessage = telegram_bot.CallbackAnswer(callbackAnswer)
		// TODO: https://core.telegram.org/bots/api#answercallbackquery, show_alert = true
		return
	},
)
