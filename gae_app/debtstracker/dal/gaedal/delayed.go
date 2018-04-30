package gaedal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"context"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/gaestandard"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/urlfetch"
)

func (UserDalGae) DelaySetUserPreferredLocale(c context.Context, delay time.Duration, userID int64, localeCode5 string) error {
	if task, err := gae.CreateDelayTask(common.QUEUE_USERS, "set-user-preferred-locale", delayedSetUserPreferredLocale, userID, localeCode5); err != nil {
		return errors.Wrap(err, "Failed to create delayed task delayedSetUserPreferredLocale")
	} else {
		task.Delay = delay
		if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			return errors.Wrap(err, "Failed to add update-users task to taskqueue")
		}
		return nil
	}
}

var delayedSetUserPreferredLocale = delay.Func("SetUserPreferredLocale", func(c context.Context, userID int64, localeCode5 string) error {
	log.Debugf(c, "delayedSetUserPreferredLocale(userID=%v, localeCode5=%v)", userID, localeCode5)
	return dal.DB.RunInTransaction(c, func(tc context.Context) error {
		user, err := dal.User.GetUserByID(tc, userID)
		if db.IsNotFound(err) {
			log.Errorf(c, "User not found by ID: %v", err)
			return nil
		}
		if err == nil && user.PreferredLanguage != localeCode5 {
			user.PreferredLanguage = localeCode5

			if err = dal.User.SaveUser(tc, user); err != nil {
				err = errors.Wrap(err, "Failed to save user to db")
			}
		}
		return err
	}, nil)
})

func (TransferDalGae) DelayUpdateTransferWithCreatorReceiptTgMessageID(c context.Context, botCode string, transferID, creatorTgChatID, creatorTgReceiptMessageID int64) error {
	//log.Debugf(c, "delayUpdateTransferWithCreatorReceiptTgMessageID(botCode=%v, transferID=%v, creatorTgChatID=%v, creatorTgReceiptMessageID=%v)", botCode, transferID, creatorTgChatID, creatorTgReceiptMessageID)

	if err := gae.CallDelayFunc(
		c, common.QUEUE_TRANSFERS, "update-transfer-with-creator-receipt-tg-message-id",
		delayedUpdateTransferWithCreatorReceiptTgMessageID,
		botCode, transferID, creatorTgChatID, creatorTgReceiptMessageID); err != nil {
		return errors.Wrap(err, "Failed to create delayed task update-transfer-with-creator-receipt-tg-message-id")
	}
	return nil
}

var delayedUpdateTransferWithCreatorReceiptTgMessageID = delay.Func("UpdateTransferWithCreatorReceiptTgMessageID", func(c context.Context, botCode string, transferID, creatorTgChatID, creatorTgReceiptMessageID int64) error {
	log.Infof(c, "delayedUpdateTransferWithCreatorReceiptTgMessageID(botCode=%v, transferID=%v, creatorTgChatID=%v, creatorReceiptTgMessageID=%v)", botCode, transferID, creatorTgChatID, creatorTgReceiptMessageID)
	return dal.DB.RunInTransaction(c, func(c context.Context) error {
		transfer, err := dal.Transfer.GetTransferByID(c, transferID)
		if err != nil {
			log.Errorf(c, "Failed to get transfer by ID: %v", err)
			if db.IsNotFound(err) {
				return nil
			} else {
				return err
			}
		}
		log.Debugf(c, "Loaded transfer: %v", transfer.TransferEntity)
		if transfer.Creator().TgBotID != botCode || transfer.Creator().TgChatID != creatorTgChatID || transfer.CreatorTgReceiptByTgMsgID != creatorTgReceiptMessageID {
			transfer.Creator().TgBotID = botCode
			transfer.Creator().TgChatID = creatorTgChatID
			transfer.CreatorTgReceiptByTgMsgID = creatorTgReceiptMessageID
			if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
				err = errors.Wrap(err, "Failed to save transfer to db")
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

func GetTelegramChatByUserID(c context.Context, userID int64) (entityID string, chat *telegram.TgChatEntityBase, err error) {
	tgChatQuery := datastore.NewQuery(telegram.ChatKind).Filter("AppUserIntID =", userID).Order("-DtUpdated")
	limit1 := 1
	tgChatQuery = tgChatQuery.Limit(limit1)
	var tgChats []*telegram.TgChatEntityBase
	tgChatKeys, err := tgChatQuery.GetAll(c, &tgChats)
	if err != nil {
		err = errors.Wrapf(err, "Failed to load telegram chat by app user id=%v", userID)
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
		err = db.NewErrNotFoundByStrID(telegram.ChatKind, "AppUserIntID="+strconv.FormatInt(userID, 10), datastore.ErrNoSuchEntity)
	}
	return
}

func DelayOnReceiptSentSuccess(c context.Context, sentAt time.Time, receiptID, transferID, tgChatID int64, tgMsgID int, tgBotID, locale string) error {
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

func DelayOnReceiptSendFail(c context.Context, receiptID, tgChatID int64, tgMsgID int, failedAt time.Time, locale, details string) error {
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

func onReceiptSentSuccess(c context.Context, sentAt time.Time, receiptID, transferID, tgChatID int64, tgMsgID int, tgBotID, locale string) (err error) {
	log.Debugf(c, "onReceiptSentSuccess(sentAt=%v, receiptID=%v, transferID=%v, tgChatID=%v, tgMsgID=%v tgBotID=%v, locale=%v)", sentAt, receiptID, transferID, tgChatID, tgMsgID, tgBotID, locale)
	if receiptID == 0 {
		return errors.New("receiptID == 0")

	}
	if transferID == 0 {
		return errors.New("transferID == 0")
	}
	var mt string
	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		receiptKey := NewReceiptKey(c, receiptID)
		transferKey := NewTransferKey(c, transferID)
		var (
			receipt        models.ReceiptEntity
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
			return errors.WithMessage(err, "failed to save transfer & receipt to datastore")
		}

		if transferEntity.DtDueOn.After(time.Now()) {
			if err := dal.Reminder.DelayCreateReminderForTransferUser(c, transferID, transferEntity.Counterparty().UserID); err != nil {
				return errors.Wrap(err, "Failed to delay creation of reminder for transfer coutnerparty")
			}
		}
		return nil
	}, dal.CrossGroupTransaction); err != nil {
		mt = err.Error()
	} else {
		var translator strongo.SingleLocaleTranslator
		if translator, err = getTranslator(c, locale); err != nil {
			return
		}
		mt = translator.Translate(trans.MESSAGE_TEXT_RECEIPT_SENT_THROW_TELEGRAM)
	}

	if err = editTgMessageText(c, tgBotID, tgChatID, tgMsgID, mt); err != nil {
		if strings.Contains(err.Error(), "Bad Request: message to edit not found") {
			log.Errorf(c, errors.WithMessage(err, fmt.Sprintf("failed to update Telegram message (tgBotID=%v, tgChatID=%v, tgMsgID=%v)",
				tgBotID, tgChatID, tgMsgID,
			)).Error())
			err = nil
		}
		return
	}

	return
}

func onReceiptSendFail(c context.Context, receiptID, tgChatID int64, tgMsgID int, failedAt time.Time, locale, details string) (err error) {
	log.Debugf(c, "onReceiptSendFail(receiptID=%v, failedAt=%v)", receiptID, failedAt)
	if receiptID == 0 {
		return errors.New("receiptID == 0")
	}
	var receipt models.Receipt
	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if receipt, err = dal.Receipt.GetReceiptByID(c, receiptID); err != nil {
			return err
		} else if receipt.DtFailed.IsZero() {
			receipt.DtFailed = failedAt
			receipt.Error = details
			if ndsErr := dal.Receipt.UpdateReceipt(c, receipt); ndsErr != nil {
				log.Errorf(c, "Failed to update Receipt with error information: %v", ndsErr) // Discard error
			}
			return err
		}
		return nil
	}, nil); err != nil {
		return
	}

	if err = editTgMessageText(c, receipt.CreatedOnID, tgChatID, tgMsgID, emoji.ERROR_ICON+" Failed to send receipt: "+details); err != nil {
		log.Errorf(c, err.Error())
		err = nil
	}
	return
}

//func getTranslatorAndTgChatID(c context.Context, userID int64) (translator strongo.SingleLocaleTranslator, tgChatID int64, err error) {
//	var (
//		//transfer models.Transfer
//		user models.AppUser
//	)
//	if user, err = dal.User.GetUserByID(c, userID); err != nil {
//		return
//	}
//	if user.TelegramUserID == 0 {
//		err = errors.New("user.TelegramUserID == 0")
//		return
//	}
//	var tgChat models.TelegramChat
//	if tgChat, err = dal.TgChat.GetTgChatByID(c, user.TelegramUserID); err != nil {
//		return
//	}
//	localeCode := tgChat.PreferredLanguage
//	if localeCode == "" {
//		localeCode = user.GetPreferredLocale()
//	}
//	if translator, err = getTranslator(c, localeCode); err != nil {
//		return
//	}
//	return
//}

func getTranslator(c context.Context, localeCode string) (translator strongo.SingleLocaleTranslator, err error) {
	log.Debugf(c, "getTranslator(localeCode=%v)", localeCode)
	var locale strongo.Locale
	if locale, err = common.TheAppContext.SupportedLocales().GetLocaleByCode5(localeCode); errors.Cause(err) == trans.ErrUnsupportedLocale {
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

func sendToTelegram(c context.Context, msg tgbotapi.Chattable, botSettings bots.BotSettings) (err error) { // TODO: Merge with same in API package
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

var delayedSendReceiptToCounterpartyByTelegram = delay.Func("dalayedSendReceiptToCounterpartyByTelegram", sendReceiptToCounterpartyByTelegram)

func updateReceiptStatus(c context.Context, receiptID int64, expectedCurrentStatus, newStatus string) (receipt models.Receipt, err error) {
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if receipt, err = dal.Receipt.GetReceiptByID(c, receiptID); err != nil {
			return
		}
		if receipt.Status != expectedCurrentStatus {
			return errReceiptStatusIsNotCreated
		}
		receipt.Status = newStatus
		if err = dal.DB.Update(c, &receipt); err != nil {
			return
		}
		return
	}, nil); err != nil {
		err = errors.WithMessage(err, fmt.Sprintf("failed to update receipt statis from %v to %v", expectedCurrentStatus, newStatus))
	}
	return
}

func sendReceiptToCounterpartyByTelegram(c context.Context, receiptID, tgChatID int64, localeCode string) (err error) {
	log.Debugf(c, "delayedSendReceiptToCounterpartyByTelegram(receiptID=%v, tgChatID=%v, localeCode=%v)", receiptID, tgChatID, localeCode)

	var receipt models.Receipt

	if receipt, err = updateReceiptStatus(c, receiptID, models.ReceiptStatusCreated, models.ReceiptStatusSending); err != nil {
		log.Errorf(c, err.Error())
		err = nil // Always stop!
		return
	}

	var transfer models.Transfer
	if transfer, err = dal.Transfer.GetTransferByID(c, receipt.TransferID); err != nil {
		log.Errorf(c, err.Error())
		if db.IsNotFound(err) {
			err = nil
			return
		}
		return
	}

	var counterpartyUser models.AppUser

	if counterpartyUser, err = dal.User.GetUserByID(c, receipt.CounterpartyUserID); err != nil {
		return
	}

	var (
		tgChat         models.TelegramChat
		failedToSend   bool
		chatsForbidden bool
	)

	creatorTgChatID, creatorTgMsgID := transfer.Creator().TgChatID, int(transfer.CreatorTgReceiptByTgMsgID)

	for _, telegramAccount := range counterpartyUser.GetTelegramAccounts() {
		if telegramAccount.App == "" {
			log.Warningf(c, "User %v has account with missing bot id => %v", counterpartyUser.ID, telegramAccount.String())
			continue
		}
		var tgChatID int64
		if tgChatID, err = strconv.ParseInt(telegramAccount.ID, 10, 64); err != nil {
			log.Errorf(c, "invalid Telegram chat ID - not an integer: %v", telegramAccount.String())
			continue
		}
		if tgChat, err = dal.TgChat.GetTgChatByID(c, telegramAccount.App, tgChatID); err != nil {
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
				log.Errorf(c, errors.Wrap(err, "Failed to call DelayOnReceiptSentSuccess()").Error())
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
			msgTextToCreator := emoji.ERROR_ICON + translator.Translate(trans.MESSAGE_TEXT_RECEIPT_NOT_SENT_AS_COUNTERPARTY_HAS_DISABLED_TG_BOT, transfer.Counterparty().ContactName)
			if err2 := DelayOnReceiptSendFail(c, receipt.ID, creatorTgChatID, creatorTgMsgID, time.Now(), translator.Locale().Code5, msgTextToCreator); err2 != nil {
				log.Errorf(c, errors.Wrap(err2, "Failed to update receipt entity with error info").Error())
			}
		}
		log.Errorf(c, "Failed to send notification to creator by Telegram (creatorTgChatID=%v, creatorTgMsgID=%v): %v", creatorTgChatID, creatorTgMsgID, err)
		msgTextToCreator := emoji.ERROR_ICON + " " + err.Error()
		if err2 := DelayOnReceiptSendFail(c, receipt.ID, creatorTgChatID, creatorTgMsgID, time.Now(), locale.Code5, msgTextToCreator); err2 != nil {
			log.Errorf(c, errors.Wrap(err2, "Failed to update receipt entity with error info").Error())
		}
		err = nil
	}

	return
}

func sendReceiptToTelegramChat(c context.Context, receipt models.Receipt, transfer models.Transfer, tgChat models.TelegramChat) (err error) {
	var messageToTranslate string
	switch transfer.Direction() {
	case models.TransferDirectionUser2Counterparty:
		messageToTranslate = trans.TELEGRAM_RECEIPT
	case models.TransferDirectionCounterparty2User:
		messageToTranslate = trans.TELEGRAM_RECEIPT
	default:
		panic(fmt.Sprintf("Unknown direction: %v", transfer.Direction()))
	}

	templateData := struct {
		FromName         string
		TransferCurrency string
	}{
		FromName:         transfer.Creator().ContactName,
		TransferCurrency: string(transfer.Currency),
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
	btnViewReceiptData := fmt.Sprintf("view-receipt?id=%v", common.EncodeID(receipt.ID)) // TODO: Pass simple digits!
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
		Text: messageText,
	}

	tgBotApi := tgbots.GetTelegramBotApiByBotCode(c, tgChat.BotID)

	if _, err = tgBotApi.Send(tgMessage); err != nil {
		return
	} else {
		log.Infof(c, "Receipt %v sent to user by Telegram bot @%v", receipt.ID, tgChat.BotID)
	}

	if receipt, err = updateReceiptStatus(c, receipt.ID, models.ReceiptStatusSending, models.ReceiptStatusSent); err != nil {
		log.Errorf(c, err.Error())
		err = nil
		return
	}
	return
}

var delayedCreateAndSendReceiptToCounterpartyByTelegram = delay.Func("delayedCreateAndSendReceiptToCounterpartyByTelegram", func(c context.Context, env strongo.Environment, transferID, toUserID int64) error {
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
		err2 := errors.Wrapf(err, "Failed to get Telegram chat for user (id=%v)", toUserID)
		if db.IsNotFound(err) {
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
	transfer, err := dal.Transfer.GetTransferByID(c, transferID)
	if err != nil {
		if db.IsNotFound(err) {
			log.Errorf(c, err.Error())
			return nil
		}
		return errors.WithMessage(err, fmt.Sprintf("Failed to get transfer by id=%v", transferID))
	}
	if localeCode == "" {
		toUser, err := dal.User.GetUserByID(c, toUserID)
		if err != nil {
			return err
		}
		localeCode = toUser.GetPreferredLocale()
	}

	var translator strongo.SingleLocaleTranslator
	if translator, err = getTranslator(c, localeCode); err != nil {
		return err
	}
	locale := translator.Locale()

	var receiptID int64
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		receipt := models.NewReceiptEntity(transfer.CreatorUserID, transferID, transfer.Counterparty().UserID, locale.Code5, telegram.PlatformID, strconv.FormatInt(tgChat.TelegramUserID, 10), general.CreatedOn{
			CreatedOnID:       transfer.Creator().TgBotID, // TODO: Replace with method call.
			CreatedOnPlatform: transfer.CreatedOnPlatform,
		})
		if receiptKey, err := gaedb.Put(c, NewReceiptIncompleteKey(c), &receipt); err != nil {
			err = errors.WithMessage(err, "failed to save receipt to DB")
		} else {
			receiptID = receiptKey.IntID()
		}
		return err
	}, nil)
	if err != nil {
		return errors.Wrapf(err, "Failed to create receipt entity")
	}
	tgChatID := (int64)(tgChat.TelegramUserID)
	if err = delaySendReceiptToCounterpartyByTelegram(c, receiptID, tgChatID, localeCode); err != nil { // TODO: ideally should be called inside transaction
		log.Errorf(c, "failed to queue receipt sending: %v", err)
		return nil
	}
	return err
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
	user, err := dal.User.GetUserByID(c, userID)
	if err != nil {
		if db.IsNotFound(err) {
			log.Errorf(c, err.Error())
			return nil
		}
		return err
	}
	if user.HasDueTransfers {
		log.Infof(c, "Already user.HasDueTransfers == %v", user.HasDueTransfers)
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
		return errors.Wrap(err, "Failed to query due reminders")
	}
	if len(keys) > 0 {
		//panic("Not implemented - refactoring in progress")
		//reminder := reminders[0]
		err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
			if user, err := dal.User.GetUserByID(tc, userID); err != nil {
				if db.IsNotFound(err) {
					log.Errorf(c, err.Error())
					return nil // Do not retry
				}
				return err
			} else if !user.HasDueTransfers {
				user.HasDueTransfers = true
				if _, err = gaedb.Put(tc, NewAppUserKey(tc, userID), user); err != nil {
					return errors.Wrap(err, "Failed to save user to db")
				}
				log.Infof(c, "User updated & saved to datastore")
			}
			return nil
		}, nil)
	}
	return err
})
