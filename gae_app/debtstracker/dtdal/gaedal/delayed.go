package gaedal

import (
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/db"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.org/asterus/debtstracker-server/gae_app/general"
	"context"
	"errors"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"github.com/strongo/app"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/gaestandard"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
	"google.golang.org/appengine/v2/delay"
	"google.golang.org/appengine/v2/urlfetch"
)

func (UserDalGae) DelaySetUserPreferredLocale(c context.Context, delay time.Duration, userID int64, localeCode5 string) error {
	if task, err := gae.CreateDelayTask(common.QUEUE_USERS, "set-user-preferred-locale", delayedSetUserPreferredLocale, userID, localeCode5); err != nil {
		return fmt.Errorf("failed to create delayed task delayedSetUserPreferredLocale: %w", err)
	} else {
		task.Delay = delay
		if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			return fmt.Errorf("failed to add update-users task to taskqueue: %w", err)
		}
		return nil
	}
}

var delayedSetUserPreferredLocale = delay.Func("SetUserPreferredLocale", func(c context.Context, userID int64, localeCode5 string) (err error) {
	log.Debugf(c, "delayedSetUserPreferredLocale(userID=%v, localeCode5=%v)", userID, localeCode5)
	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	return db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) error {
		user, err := facade.User.GetUserByID(tc, userID)
		if dal.IsNotFound(err) {
			log.Errorf(c, "User not found by ID: %v", err)
			return nil
		}
		if err == nil && user.Data.PreferredLanguage != localeCode5 {
			user.Data.PreferredLanguage = localeCode5

			if err = facade.User.SaveUser(tc, tx, user); err != nil {
				err = fmt.Errorf("failed to save user to db: %w", err)
			}
		}
		return err
	}, nil)
})

func (TransferDalGae) DelayUpdateTransferWithCreatorReceiptTgMessageID(c context.Context, botCode string, transferID, creatorTgChatID, creatorTgReceiptMessageID int64) error {
	// log.Debugf(c, "delayUpdateTransferWithCreatorReceiptTgMessageID(botCode=%v, transferID=%v, creatorTgChatID=%v, creatorTgReceiptMessageID=%v)", botCode, transferID, creatorTgChatID, creatorTgReceiptMessageID)

	if err := gae.CallDelayFunc(
		c, common.QUEUE_TRANSFERS, "update-transfer-with-creator-receipt-tg-message-id",
		delayedUpdateTransferWithCreatorReceiptTgMessageID,
		botCode, transferID, creatorTgChatID, creatorTgReceiptMessageID); err != nil {
		return fmt.Errorf("failed to create delayed task update-transfer-with-creator-receipt-tg-message-id: %w", err)
	}
	return nil
}

var delayedUpdateTransferWithCreatorReceiptTgMessageID = delay.Func("UpdateTransferWithCreatorReceiptTgMessageID", func(c context.Context, botCode string, transferID int, creatorTgChatID, creatorTgReceiptMessageID int64) (err error) {
	log.Infof(c, "delayedUpdateTransferWithCreatorReceiptTgMessageID(botCode=%v, transferID=%v, creatorTgChatID=%v, creatorReceiptTgMessageID=%v)", botCode, transferID, creatorTgChatID, creatorTgReceiptMessageID)
	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	return db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) error {
		transfer, err := facade.Transfers.GetTransferByID(c, tx, transferID)
		if err != nil {
			log.Errorf(c, "Failed to get transfer by ID: %v", err)
			if dal.IsNotFound(err) {
				return nil
			} else {
				return err
			}
		}
		log.Debugf(c, "Loaded transfer: %v", transfer.Data)
		if transfer.Data.Creator().TgBotID != botCode || transfer.Data.Creator().TgChatID != creatorTgChatID || transfer.Data.CreatorTgReceiptByTgMsgID != creatorTgReceiptMessageID {
			transfer.Data.Creator().TgBotID = botCode
			transfer.Data.Creator().TgChatID = creatorTgChatID
			transfer.Data.CreatorTgReceiptByTgMsgID = creatorTgReceiptMessageID
			if err = facade.Transfers.SaveTransfer(c, tx, transfer); err != nil {
				err = fmt.Errorf("failed to save transfer to db: %w", err)
			}
		}
		return err
	}, nil)
})

func (ReceiptDalGae) DelayCreateAndSendReceiptToCounterpartyByTelegram(c context.Context, env strongo.Environment, transferID, userID int64) error {
	log.Debugf(c, "delaySendReceiptToCounterpartyByTelegram(env=%v, transferID=%v, userID=%v)", env, transferID, userID)

	if task, err := gae.CreateDelayTask(common.QUEUE_RECEIPTS, "create-and-send-receipt-for-counterparty-by-telegram", delayedCreateAndSendReceiptToCounterpartyByTelegram, env, transferID, userID); err != nil {
		return err
	} else {
		task.Delay = time.Duration(1 * time.Second / 10)
		if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_RECEIPTS); err != nil {
			return err
		}
	}
	return nil
}

func GetTelegramChatByUserID(c context.Context, userID int64) (entityID string, chat *tgstore.ChatEntity, err error) {
	tgChatQuery := datastore.NewQuery(tgstore.TgChatCollection).Filter("AppUserIntID =", userID).Order("-DtUpdated")
	limit1 := 1
	tgChatQuery = tgChatQuery.Limit(limit1)
	var tgChats []*tgstore.ChatEntity
	tgChatKeys, err := tgChatQuery.GetAll(c, &tgChats)
	if err != nil {
		err = fmt.Errorf("failed to load telegram chat by app user id=%v: %w", userID, err)
		return
	}
	if len(tgChatKeys) == limit1 {
		if entityID = tgChatKeys[0].StringID(); entityID == "" {
			entityID = strconv.FormatInt(tgChatKeys[0].IntID(), 10)
		}
		chat = tgChats[0]
		return
	} else {
		log.Debugf(c, "len(tgChatKeys): %v", len(tgChatKeys))
		err = db.NewErrNotFoundByStrID(tgstore.TgChatCollection, "AppUserIntID="+strconv.FormatInt(userID, 10), datastore.ErrNoSuchEntity)
	}
	return
}

func DelayOnReceiptSentSuccess(c context.Context, sentAt time.Time, receiptID, transferID int, tgChatID int64, tgMsgID int, tgBotID, locale string) error {
	if receiptID == 0 {
		return errors.New("receiptID == 0")
	}
	if transferID == 0 {
		return errors.New("transferID == 0")
	}
	if err := gae.CallDelayFunc(c, common.QUEUE_RECEIPTS, "on-receipt-sent-success", delayedOnReceiptSentSuccess, sentAt, receiptID, transferID, tgChatID, tgMsgID, tgBotID, locale); err != nil {
		log.Errorf(c, err.Error())
		return onReceiptSentSuccess(c, sentAt, receiptID, transferID, tgChatID, tgMsgID, tgBotID, locale)
	}
	return nil
}

func DelayOnReceiptSendFail(c context.Context, receiptID int, tgChatID int64, tgMsgID int, failedAt time.Time, locale, details string) error {
	if receiptID == 0 {
		return errors.New("receiptID == 0")
	}
	if failedAt.IsZero() {
		return errors.New("failedAt.IsZero()")
	}
	if err := gae.CallDelayFunc(c, common.QUEUE_RECEIPTS, "on-receipt-send-fail", delayedOnReceiptSendFail, receiptID, tgChatID, tgMsgID, failedAt, locale, details); err != nil {
		log.Errorf(c, err.Error())
		return onReceiptSendFail(c, receiptID, tgChatID, tgMsgID, failedAt, locale, details)
	}
	return nil
}

var delayedOnReceiptSentSuccess = delay.Func("onReceiptSentSuccess", onReceiptSentSuccess)
var delayedOnReceiptSendFail = delay.Func("onReceiptSendFail", onReceiptSendFail)

func onReceiptSentSuccess(c context.Context, sentAt time.Time, receiptID, transferID int, tgChatID int64, tgMsgID int, tgBotID, locale string) (err error) {
	log.Debugf(c, "onReceiptSentSuccess(sentAt=%v, receiptID=%v, transferID=%v, tgChatID=%v, tgMsgID=%v tgBotID=%v, locale=%v)", sentAt, receiptID, transferID, tgChatID, tgMsgID, tgBotID, locale)
	if receiptID == 0 {
		log.Errorf(c, "receiptID == 0")
		return

	}
	if transferID == 0 {
		log.Errorf(c, "transferID == 0")
		return
	}
	if tgChatID == 0 {
		log.Errorf(c, "tgChatID == 0")
		return
	}
	if tgMsgID == 0 {
		log.Errorf(c, "tgMsgID == 0")
		return
	}
	var mt string
	var receipt models.ReceiptEntity
	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}
	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		receiptKey := NewReceiptKey(c, receiptID)
		transferKey := NewTransferKey(c, transferID)
		var (
			transferEntity models.TransferEntity
		)
		// TODO: Replace with DAL call?
		if err := gaedb.GetMulti(c, []*datastore.Key{receiptKey, transferKey}, []interface{}{&receipt, &transferEntity}); err != nil {
			return err
		}
		if receipt.TransferID != transferID {
			return errors.New("receipt.TransferID != transferID")
		}
		if receipt.Status == models.ReceiptStatusSent {
			return nil
		}

		transferEntity.Counterparty().TgBotID = tgBotID
		transferEntity.Counterparty().TgChatID = tgChatID
		receipt.DtSent = sentAt
		receipt.Status = models.ReceiptStatusSent
		if _, err := gaedb.PutMulti(c, []*datastore.Key{transferKey, receiptKey}, []interface{}{&transferEntity, &receipt}); err != nil {
			return fmt.Errorf("failed to save transfer & receipt to datastore: %w", err)
		}

		if transferEntity.DtDueOn.After(time.Now()) {
			if err := dtdal.Reminder.DelayCreateReminderForTransferUser(c, transferID, transferEntity.Counterparty().UserID); err != nil {
				return fmt.Errorf("failed to delay creation of reminder for transfer counterparty: %w", err)
			}
		}
		return nil
	}); err != nil {
		mt = err.Error()
	} else {
		var translator strongo.SingleLocaleTranslator
		if translator, err = getTranslator(c, locale); err != nil {
			return
		}
		mt = translator.Translate(trans.MESSAGE_TEXT_RECEIPT_SENT_THROW_TELEGRAM)
	}

	if err = editTgMessageText(c, tgBotID, tgChatID, tgMsgID, mt); err != nil {
		errMessage := err.Error()
		err = fmt.Errorf("failed to update Telegram message (botID=%v, chatID=%v, msgID=%v): %w", tgBotID, tgChatID, tgMsgID, err)
		if strings.Contains(errMessage, "Bad Request") && strings.Contains(errMessage, " not found") {
			logMessage := log.Errorf
			switch {
			case receipt.DtCreated.Before(time.Now().Add(-time.Hour * 24)):
				logMessage = log.Debugf
			case receipt.DtCreated.Before(time.Now().Add(-time.Hour)):
				logMessage = log.Infof
			case receipt.DtCreated.Before(time.Now().Add(-time.Minute)):
				logMessage = log.Warningf
			}
			logMessage(c, err.Error())
			err = nil
		}
		return
	}
	return
}

func onReceiptSendFail(c context.Context, receiptID int, tgChatID int64, tgMsgID int, failedAt time.Time, locale, details string) (err error) {
	log.Debugf(c, "onReceiptSendFail(receiptID=%v, failedAt=%v)", receiptID, failedAt)
	if receiptID == 0 {
		return errors.New("receiptID == 0")
	}
	var receipt models.Receipt
	if err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
		if receipt, err = dtdal.Receipt.GetReceiptByID(c, receiptID); err != nil {
			return err
		} else if receipt.Data.DtFailed.IsZero() {
			receipt.Data.DtFailed = failedAt
			receipt.Data.Error = details
			if ndsErr := dtdal.Receipt.UpdateReceipt(c, receipt); ndsErr != nil {
				log.Errorf(c, "Failed to update Receipt with error information: %v", ndsErr) // Discard error
			}
			return err
		}
		return nil
	}, nil); err != nil {
		return
	}

	if err = editTgMessageText(c, receipt.Data.CreatedOnID, tgChatID, tgMsgID, emoji.ERROR_ICON+" Failed to send receipt: "+details); err != nil {
		log.Errorf(c, err.Error())
		err = nil
	}
	return
}

// func getTranslatorAndTgChatID(c context.Context, userID int64) (translator strongo.SingleLocaleTranslator, tgChatID int64, err error) {
// 	var (
// 		//transfer models.Transfer
// 		user models.AppUser
// 	)
// 	if user, err = facade.User.GetUserByID(c, userID); err != nil {
// 		return
// 	}
// 	if user.TelegramUserID == 0 {
// 		err = errors.New("user.TelegramUserID == 0")
// 		return
// 	}
// 	var tgChat models.TelegramChat
// 	if tgChat, err = dtdal.TgChat.GetTgChatByID(c, user.TelegramUserID); err != nil {
// 		return
// 	}
// 	localeCode := tgChat.PreferredLanguage
// 	if localeCode == "" {
// 		localeCode = user.GetPreferredLocale()
// 	}
// 	if translator, err = getTranslator(c, localeCode); err != nil {
// 		return
// 	}
// 	return
// }

func getTranslator(c context.Context, localeCode string) (translator strongo.SingleLocaleTranslator, err error) {
	log.Debugf(c, "getTranslator(localeCode=%v)", localeCode)
	var locale strongo.Locale
	if locale, err = common.TheAppContext.SupportedLocales().GetLocaleByCode5(localeCode); errors.Is(err, trans.ErrUnsupportedLocale) {
		localeCode = strongo.LocaleCodeEnUS
	}
	if locale, err = common.TheAppContext.SupportedLocales().GetLocaleByCode5(localeCode); err != nil {
		return
	}
	translator = strongo.NewSingleMapTranslator(locale, common.TheAppContext.GetTranslator(c))
	return
}

func editTgMessageText(c context.Context, tgBotID string, tgChatID int64, tgMsgID int, text string) (err error) {
	msg := tgbotapi.NewEditMessageText(tgChatID, tgMsgID, "", text)
	telegramBots := tgbots.Bots(gaestandard.GetEnvironment(c), nil)
	botSettings, ok := telegramBots.ByCode[tgBotID]
	if !ok {
		return fmt.Errorf("Bot settings not found by tgChat.BotID=%v, out of %v items", tgBotID, len(telegramBots.ByCode))
	}
	if err = sendToTelegram(c, msg, botSettings); err != nil {
		return
	}
	return
}

func sendToTelegram(c context.Context, msg tgbotapi.Chattable, botSettings botsfw.BotSettings) (err error) { // TODO: Merge with same in API package
	tgApi := tgbotapi.NewBotAPIWithClient(botSettings.Token, urlfetch.Client(c))
	if _, err = tgApi.Send(msg); err != nil {
		return
	}
	return
}

var errReceiptStatusIsNotCreated = errors.New("receipt is not in 'created' status")

func delaySendReceiptToCounterpartyByTelegram(c context.Context, receiptID, tgChatID int64, localeCode string) error {
	if task, err := gae.CreateDelayTask(common.QUEUE_RECEIPTS, "send-receipt-to-counterparty-by-telegram", delayedSendReceiptToCounterpartyByTelegram, receiptID, tgChatID, localeCode); err != nil {
		return err
	} else {
		task.Delay = time.Duration(1 * time.Second / 10)
		if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_RECEIPTS); err != nil {
			return err
		}
	}
	return nil
}

var delayedSendReceiptToCounterpartyByTelegram = delay.Func("delayedSendReceiptToCounterpartyByTelegram", sendReceiptToCounterpartyByTelegram)

func updateReceiptStatus(c context.Context, tx dal.ReadwriteTransaction, receiptID int, expectedCurrentStatus, newStatus string) (receipt models.Receipt, err error) {

	if err = func() (err error) {
		if receipt, err = dtdal.Receipt.GetReceiptByID(c, receiptID); err != nil {
			return
		}
		if receipt.Data.Status != expectedCurrentStatus {
			return errReceiptStatusIsNotCreated
		}
		receipt.Data.Status = newStatus
		if err = tx.Set(c, receipt.Record); err != nil {
			return
		}
		return
	}(); err != nil {
		err = fmt.Errorf("failed to update receipt status from %v to %v: %w", expectedCurrentStatus, newStatus, err)
	}
	return
}

func sendReceiptToCounterpartyByTelegram(c context.Context, receiptID int, tgChatID int64, localeCode string) (err error) {
	log.Debugf(c, "delayedSendReceiptToCounterpartyByTelegram(receiptID=%v, tgChatID=%v, localeCode=%v)", receiptID, tgChatID, localeCode)

	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}
	if err := db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
		var receipt models.Receipt

		if receipt, err = updateReceiptStatus(c, tx, receiptID, models.ReceiptStatusCreated, models.ReceiptStatusSending); err != nil {
			log.Errorf(c, err.Error())
			err = nil // Always stop!
			return
		}

		var transfer models.Transfer
		if transfer, err = facade.Transfers.GetTransferByID(c, tx, receipt.Data.TransferID); err != nil {
			log.Errorf(c, err.Error())
			if dal.IsNotFound(err) {
				err = nil
				return
			}
			return
		}

		var counterpartyUser models.AppUser

		if counterpartyUser, err = facade.User.GetUserByID(c, receipt.Data.CounterpartyUserID); err != nil {
			return
		}

		var (
			tgChat         models.TelegramChat
			failedToSend   bool
			chatsForbidden bool
		)

		creatorTgChatID, creatorTgMsgID := transfer.Data.Creator().TgChatID, int(transfer.Data.CreatorTgReceiptByTgMsgID)

		for _, telegramAccount := range counterpartyUser.Data.GetTelegramAccounts() {
			if telegramAccount.App == "" {
				log.Warningf(c, "User %v has account with missing bot id => %v", counterpartyUser.ID, telegramAccount.String())
				continue
			}
			var tgChatID int64
			if tgChatID, err = strconv.ParseInt(telegramAccount.ID, 10, 64); err != nil {
				log.Errorf(c, "invalid Telegram chat ID - not an integer: %v", telegramAccount.String())
				continue
			}
			if tgChat, err = dtdal.TgChat.GetTgChatByID(c, telegramAccount.App, tgChatID); err != nil {
				log.Errorf(c, "failed to load user's Telegram chat entity: %v", err)
				continue
			}
			if tgChat.DtForbiddenLast.IsZero() {
				if err = sendReceiptToTelegramChat(c, receipt, transfer, tgChat); err != nil {
					failedToSend = true
					if _, forbidden := err.(tgbotapi.ErrAPIForbidden); forbidden || strings.Contains(err.Error(), "Bad Request: chat not found") {
						chatsForbidden = true
						log.Infof(c, "Telegram chat not found or disabled (%v): %v", tgChat.ID, err)
						if err2 := gaehost.MarkTelegramChatAsForbidden(c, tgChat.BotID, tgChat.TelegramUserID, time.Now()); err2 != nil {
							log.Errorf(c, "Failed to call MarkTelegramChatAsStopped(): %v", err2.Error())
						}
						return nil
					}
					return
				}
				if err = DelayOnReceiptSentSuccess(c, time.Now(), receipt.ID, transfer.ID, creatorTgChatID, creatorTgMsgID, tgChat.BotID, localeCode); err != nil {
					log.Errorf(c, fmt.Errorf("failed to call DelayOnReceiptSentSuccess(): %w", err).Error())
				}
				return
			} else {
				log.Debugf(c, "tgChat is forbidden: %v", telegramAccount.String())
			}
			break
		}

		if failedToSend { // Notify creator that receipt has not been sent
			var translator strongo.SingleLocaleTranslator
			if translator, err = getTranslator(c, localeCode); err != nil {
				return err
			}

			locale := translator.Locale()
			if chatsForbidden {
				msgTextToCreator := emoji.ERROR_ICON + translator.Translate(trans.MESSAGE_TEXT_RECEIPT_NOT_SENT_AS_COUNTERPARTY_HAS_DISABLED_TG_BOT, transfer.Data.Counterparty().ContactName)
				if err2 := DelayOnReceiptSendFail(c, receipt.ID, creatorTgChatID, creatorTgMsgID, time.Now(), translator.Locale().Code5, msgTextToCreator); err2 != nil {
					log.Errorf(c, fmt.Errorf("failed to update receipt entity with error info: %w", err2).Error())
				}
			}
			log.Errorf(c, "Failed to send notification to creator by Telegram (creatorTgChatID=%v, creatorTgMsgID=%v): %v", creatorTgChatID, creatorTgMsgID, err)
			msgTextToCreator := emoji.ERROR_ICON + " " + err.Error()
			if err2 := DelayOnReceiptSendFail(c, receipt.ID, creatorTgChatID, creatorTgMsgID, time.Now(), locale.Code5, msgTextToCreator); err2 != nil {
				log.Errorf(c, fmt.Errorf("failed to update receipt entity with error info: %w", err2).Error())
			}
			err = nil
		}
		return err
	}); err != nil {
		return
	}
	return
}

func sendReceiptToTelegramChat(c context.Context, receipt models.Receipt, transfer models.Transfer, tgChat models.TelegramChat) (err error) {
	var messageToTranslate string
	switch transfer.Data.Direction() {
	case models.TransferDirectionUser2Counterparty:
		messageToTranslate = trans.TELEGRAM_RECEIPT
	case models.TransferDirectionCounterparty2User:
		messageToTranslate = trans.TELEGRAM_RECEIPT
	default:
		panic(fmt.Errorf("Unknown direction: %v", transfer.Data.Direction()))
	}

	templateData := struct {
		FromName         string
		TransferCurrency string
	}{
		FromName:         transfer.Data.Creator().ContactName,
		TransferCurrency: string(transfer.Data.Currency),
	}

	var translator strongo.SingleLocaleTranslator
	if translator, err = getTranslator(c, tgChat.GetPreferredLanguage()); err != nil {
		return err
	}

	messageText, err := common.TextTemplates.RenderTemplate(c, translator, messageToTranslate, templateData)
	if err != nil {
		return err
	}
	messageText = emoji.INCOMING_ENVELOP_ICON + " " + messageText

	log.Debugf(c, "Message: %v", messageText)

	btnViewReceiptText := emoji.CLIPBOARD_ICON + " " + translator.Translate(trans.BUTTON_TEXT_SEE_RECEIPT_DETAILS)
	btnViewReceiptData := fmt.Sprintf("view-receipt?id=%v", common.EncodeID(int64(receipt.ID))) // TODO: Pass simple digits!
	tgMessage := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: tgChat.TelegramUserID,
			ReplyMarkup: tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnViewReceiptText, btnViewReceiptData)),
				},
			},
		},
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
		Text:                  messageText,
	}

	tgBotApi := tgbots.GetTelegramBotApiByBotCode(c, tgChat.BotID)

	if _, err = tgBotApi.Send(tgMessage); err != nil {
		return
	} else {
		log.Infof(c, "Receipt %v sent to user by Telegram bot @%v", receipt.ID, tgChat.BotID)
	}

	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return err
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
		if receipt, err = updateReceiptStatus(c, tx, receipt.ID, models.ReceiptStatusSending, models.ReceiptStatusSent); err != nil {
			log.Errorf(c, err.Error())
			err = nil
			return
		}
		return err
	})
	return
}

var delayedCreateAndSendReceiptToCounterpartyByTelegram = delay.Func("delayedCreateAndSendReceiptToCounterpartyByTelegram", func(c context.Context, env strongo.Environment, transferID int, toUserID int64) error {
	log.Debugf(c, "delayedCreateAndSendReceiptToCounterpartyByTelegram(transferID=%v, toUserID=%v)", transferID, toUserID)
	if transferID == 0 {
		log.Errorf(c, "transferID == 0")
		return nil
	}
	if toUserID == 0 {
		log.Errorf(c, "toUserID == 0")
		return nil
	}
	chatEntityID, tgChat, err := GetTelegramChatByUserID(c, toUserID)
	if err != nil {
		err2 := fmt.Errorf("failed to get Telegram chat for user (id=%v): %w", toUserID, err)
		if dal.IsNotFound(err) {
			log.Infof(c, "No telegram for user or user not found")
			return nil
		} else {
			return err2
		}
	}
	if chatEntityID == "" {
		log.Infof(c, "No telegram for user")
		return nil
	}
	localeCode := tgChat.PreferredLanguage
	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return err
	}
	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		transfer, err := facade.Transfers.GetTransferByID(c, tx, transferID)
		if err != nil {
			if dal.IsNotFound(err) {
				log.Errorf(c, err.Error())
				return nil
			}
			return fmt.Errorf("failed to get transfer by id=%v: %v", transferID, err)
		}
		if localeCode == "" {
			toUser, err := facade.User.GetUserByID(c, toUserID)
			if err != nil {
				return err
			}
			localeCode = toUser.Data.GetPreferredLocale()
		}

		var translator strongo.SingleLocaleTranslator
		if translator, err = getTranslator(c, localeCode); err != nil {
			return err
		}
		locale := translator.Locale()

		var receiptID int64
		err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
			receipt := models.NewReceiptEntity(transfer.Data.CreatorUserID, transferID, transfer.Data.Counterparty().UserID, locale.Code5, telegram.PlatformID, strconv.FormatInt(tgChat.TelegramUserID, 10), general.CreatedOn{
				CreatedOnID:       transfer.Data.Creator().TgBotID, // TODO: Replace with method call.
				CreatedOnPlatform: transfer.Data.CreatedOnPlatform,
			})
			if receiptKey, err := gaedb.Put(c, NewReceiptIncompleteKey(c), &receipt); err != nil {
				err = fmt.Errorf("failed to save receipt to DB: %w", err)
			} else {
				receiptID = receiptKey.IntID()
			}
			return err
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to create receipt entity: %w", err)
		}
		tgChatID := (int64)(tgChat.TelegramUserID)
		if err = delaySendReceiptToCounterpartyByTelegram(c, receiptID, tgChatID, localeCode); err != nil { // TODO: ideally should be called inside transaction
			log.Errorf(c, "failed to queue receipt sending: %v", err)
			return nil
		}
		return err
	}); err != nil {
		return err
	}
	return nil
})

func (UserDalGae) DelayUpdateUserHasDueTransfers(c context.Context, userID int64) error {
	if userID == 0 {
		panic("userID == 0")
	}
	return gae.CallDelayFunc(c, common.QUEUE_USERS, "update-user-has-due-transfers", delayedUpdateUserHasDueTransfers, userID)
}

var delayedUpdateUserHasDueTransfers = delay.Func("delayedUpdateUserHasDueTransfers", func(c context.Context, userID int64) (err error) {
	log.Debugf(c, "delayedUpdateUserHasDueTransfers(userID=%v)", userID)
	if userID == 0 {
		log.Errorf(c, "userID == 0")
		return nil
	}
	user, err := facade.User.GetUserByID(c, userID)
	if err != nil {
		if dal.IsNotFound(err) {
			log.Errorf(c, err.Error())
			return nil
		}
		return err
	}
	if user.Data.HasDueTransfers {
		log.Infof(c, "Already user.HasDueTransfers == %v", user.Data.HasDueTransfers)
		return nil
	}

	q := datastore.NewQuery(models.TransferKind)
	q = q.Filter("BothUserIDs =", userID)
	q = q.Filter("IsOutstanding =", true)
	q = q.Filter("DtDueOn >", time.Time{})
	q = q.KeysOnly()
	q = q.Limit(1)
	var keys []*datastore.Key
	if _, err = q.GetAll(c, nil); err != nil {
		return fmt.Errorf("failed to query due reminders: %w", err)
	}
	if len(keys) > 0 {
		// panic("Not implemented - refactoring in progress")
		// reminder := reminders[0]
		err = dtdal.DB.RunInTransaction(c, func(tc context.Context) error {
			if user, err := facade.User.GetUserByID(tc, userID); err != nil {
				if dal.IsNotFound(err) {
					log.Errorf(c, err.Error())
					return nil // Do not retry
				}
				return err
			} else if !user.Data.HasDueTransfers {
				user.Data.HasDueTransfers = true
				if _, err = gaedb.Put(tc, NewAppUserKey(tc, userID), user); err != nil {
					return fmt.Errorf("failed to save user to db: %w", err)
				}
				log.Infof(c, "User updated & saved to datastore")
			}
			return nil
		}, nil)
	}
	return err
})
