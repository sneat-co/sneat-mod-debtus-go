package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.com/asterus/debtstracker-server/gae_app/gaestandard"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/urlfetch"
	"strconv"
	"time"
)

func (_ UserDalGae) DelaySetUserPreferredLocale(c context.Context, delay time.Duration, userID int64, localeCode5 string) error {
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

func (_ TransferDalGae) DelayUpdateTransferWithCreatorReceiptTgMessageID(c context.Context, botCode string, transferID, creatorTgChatID, creatorTgReceiptMessageID int64) error {
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

func (_ ReceiptDalGae) DelaySendReceiptToCounterpartyByTelegram(c context.Context, env strongo.Environment, transferID, userID int64) error {
	log.Debugf(c, "delaySendReceiptToCounterpartyByTelegram(env=%v, transferID=%v, userID=%v)", env, transferID, userID)

	queueName := common.QUEUE_RECEIPTS
	if task, err := gae.CreateDelayTask(queueName, "send-receipt-to-counterparty-by-telegram", delayedSendReceiptToCounterpartyByTelegram, env, transferID, userID); err != nil {
		return errors.Wrapf(err, "Failed to create delayed task")
	} else {
		task.Delay = time.Duration(1 * time.Second)
		if _, err = gae.AddTaskToQueue(c, task, queueName); err != nil {
			return errors.Wrapf(err, "Failed to add delayed task to nitifcation queue")
		}
		log.Debugf(c, "Queued for execution: delayedSendReceiptToCounterpartyByTelegram()")
	}
	return nil
}

func GetTelegramChatByUserID(c context.Context, userID int64) (entityID string, chat *telegram_bot.TelegramChatEntityBase, err error) {
	tgChatQuery := datastore.NewQuery(telegram_bot.TelegramChatKind).Filter("AppUserIntID =", userID).Order("-DtUpdated")
	limit1 := 1
	tgChatQuery = tgChatQuery.Limit(limit1)
	var tgChats []*telegram_bot.TelegramChatEntityBase
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
		err = db.NewErrNotFoundByStrID(telegram_bot.TelegramChatKind, "AppUserIntID="+strconv.FormatInt(userID, 10), datastore.ErrNoSuchEntity)
	}
	return
}

func DelayOnReceiptSentSuccess(c context.Context, sentAt time.Time, receiptID, transferID, tgChatID int64, tgBotID string) error {
	if receiptID == 0 {
		return errors.New("receiptID == 0")
	}
	if transferID == 0 {
		return errors.New("transferID == 0")
	}
	if err := gae.CallDelayFunc(c, common.QUEUE_RECEIPTS, "on-receipt-sent-success", delayedOnReceiptSentSuccess, sentAt, receiptID, transferID, tgChatID, tgBotID); err != nil {
		log.Errorf(c, err.Error())
		return onReceiptSentSuccess(c, sentAt, receiptID, transferID, tgChatID, tgBotID)
	}
	return nil
}

func DelayOnReceiptSendFail(c context.Context, receiptID int64, failedAt time.Time, details string) error {
	if receiptID == 0 {
		return errors.New("receiptID == 0")
	}
	if failedAt.IsZero() {
		return errors.New("failedAt.IsZero()")
	}
	if err := gae.CallDelayFunc(c, common.QUEUE_RECEIPTS, "on-receipt-send-fail", delayedOnReceiptSendFail, receiptID, failedAt, details); err != nil {
		log.Errorf(c, err.Error())
		return onReceiptSendFail(c, receiptID, failedAt, details)
	}
	return nil
}

var delayedOnReceiptSentSuccess = delay.Func("onReceiptSentSuccess", onReceiptSentSuccess)
var delayedOnReceiptSendFail = delay.Func("onReceiptSendFail", onReceiptSendFail)

func onReceiptSentSuccess(c context.Context, sentAt time.Time, receiptID, transferID, tgChatID int64, tgBotID string) error {
	log.Debugf(c, "onReceiptSentSuccess(sentAt=%v, receiptID=%v, transferID=%v, tgBotID=%v)", sentAt, receiptID, transferID, tgBotID)
	if receiptID == 0 {
		return errors.New("receiptID == 0")

	}
	if transferID == 0 {
		return errors.New("transferID == 0")
	}
	return dal.DB.RunInTransaction(c, func(c context.Context) error {
		receiptKey := NewReceiptKey(c, receiptID)
		transferKey := NewTransferKey(c, transferID)
		var (
			receipt        models.ReceiptEntity
			transferEntity models.TransferEntity
		)
		if err := gaedb.GetMulti(c, []*datastore.Key{receiptKey, transferKey}, []interface{}{&receipt, &transferEntity}); err != nil {
			return err
		}
		if receipt.TransferID != transferID {
			return errors.New("receipt.TransferID != transferID")
		}

		transferEntity.Counterparty().TgBotID = tgBotID
		transferEntity.Counterparty().TgChatID = tgChatID
		receipt.DtSent = sentAt
		receipt.Status = models.ReceiptStatusSent
		if _, err := gaedb.PutMulti(c, []*datastore.Key{transferKey, receiptKey}, []interface{}{&transferEntity, &receipt}); err != nil {
			return errors.Wrap(err, "Failed to save transfer & receipt to datastore")
		}

		if transferEntity.DtDueOn.After(time.Now()) {
			if err := dal.Reminder.DelayCreateReminderForTransferUser(c, transferID, transferEntity.Counterparty().UserID); err != nil {
				return errors.Wrap(err, "Failed to delay creation of reminder for transfer coutnerparty")
			}
		}
		return nil
	}, dal.CrossGroupTransaction)
}

func onReceiptSendFail(c context.Context, receiptID int64, failedAt time.Time, details string) error {
	log.Debugf(c, "onReceiptSendFail(receiptID=%v, failedAt=%v)", receiptID, failedAt)
	if receiptID == 0 {
		return errors.New("receiptID == 0")
	}
	return dal.DB.RunInTransaction(c, func(c context.Context) error {
		if receipt, err := dal.Receipt.GetReceiptByID(c, receiptID); err != nil {
			return err
		} else {
			receipt.DtFailed = failedAt
			receipt.Error = details
			if _, ndsErr := gaedb.Put(c, NewReceiptKey(c, receiptID), receipt); ndsErr != nil {
				log.Errorf(c, "Failed to update Receipt with error information: %v", ndsErr) // Discard error
			}
			return err
		}
	}, nil)
}

var delayedSendReceiptToCounterpartyByTelegram = delay.Func("sendReceiptToCounterparty", func(c context.Context, env strongo.Environment, transferID, toUserID int64) error {
	log.Debugf(c, "delayedSendReceiptToCounterpartyByTelegram(transferID=%v, toUserID=%v)", transferID, toUserID)
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
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(c, "Transfer not found by id=%v", transferID)
			return nil
		}
		return errors.Wrapf(err, "Failed to get transfer by id=%v", transferID)
	}
	if localeCode == "" {
		toUser, err := dal.User.GetUserByID(c, toUserID)
		if err != nil {
			return err
		}
		localeCode = toUser.PreferredLocale()
	}

	toUserBotSettings, err := telegram.GetBotSettingsByLang(gaestandard.GetEnvironment(c), bot.ProfileDebtus, localeCode)
	if err != nil {
		panic(errors.Wrap(err, "Bot settings not found by locale").Error())
	}

	translator := common.TheAppContext.GetTranslator(c)

	locale, err := common.TheAppContext.SupportedLocales().GetLocaleByCode5(localeCode)
	if err != nil {
		return errors.Wrapf(err, "Failed to get locale by code5 (%v)", localeCode)
	}
	//contactDalGae, transfer, err := GetTransferByID(c, transferID)
	//if err != nil {
	//	log.Errorf(c, "Failed to get transfer (%v): %v", transferID, err)
	//	return err
	//}
	//counterpartyUser, err := GetUserByID(c, transfer.Counterparty().UserID)
	//if err != nil {
	//	log.Errorf(c, "Failed to get counterparty user (%v): %v", transfer.Counterparty().UserID, err)
	//	return err
	//}
	//locale, ok := trans.SupportedLocalesByCode5[counterpartyUser.PreferredLocale()]
	//if !ok {
	//	creatorUser, err := GetUserByID(c, transfer.CreatorUserID)
	//	if err != nil {
	//		log.Errorf(c, "Failed to get creator user (%v): %v", transfer.CreatorUserID, err)
	//		return err
	//	}
	//	locale, ok = trans.SupportedLocalesByCode5[creatorUser.PreferredLocale()]
	//	if !ok {
	//		locale = strongo.LocaleEnUs
	//	}
	//}

	var receiptID int64
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		receipt := models.NewReceiptEntity(transfer.CreatorUserID, transferID, transfer.Counterparty().UserID, locale.Code5, telegram_bot.TelegramPlatformID, strconv.Itoa(tgChat.TelegramUserID), general.CreatedOn{
			CreatedOnID:       transfer.Creator().TgBotID, // TODO: Replace with method call.
			CreatedOnPlatform: transfer.CreatedOnPlatform,
		})
		if receiptKey, err := gaedb.Put(c, NewReceiptIncompleteKey(c), &receipt); err != nil {
			return errors.Wrap(err, "Failed to save receipt to DB")
		} else {
			receiptID = receiptKey.IntID()
			return nil
		}
	}, nil)

	if err != nil {
		return errors.Wrapf(err, "Failed to create receipt entity")
	}
	log.Infof(c, "Receipt id=%v", receiptID)
	/* Send receipt */
	{
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

		messageText, err := common.TextTemplates.RenderTemplate(c, strongo.NewSingleMapTranslator(locale, translator), messageToTranslate, templateData)
		if err != nil {
			return err
		}
		messageText = emoji.INCOMING_ENVELOP_ICON + " " + messageText

		//utmParams := UtmParams{
		//	Source: "telegram",
		//	Medium: "bot",
		//	Campaign: UTM_CAMPAIGN_RECEIPT,
		//}
		//receiptUrl :=  GetReceiptUrl(locale, transferID, utmParams), translator.Translate(trans.COMMAND_TEXT_SEE_RECEIPT_DETAILS, locale.Code5)

		log.Debugf(c, "Message: %v", messageText)

		tgChatID := (int64)(tgChat.TelegramUserID)

		btnViewReceiptText := emoji.CLIPBOARD_ICON + " " + translator.Translate(trans.BUTTON_TEXT_SEE_RECEIPT_DETAILS, locale.Code5)
		btnViewReceiptData := fmt.Sprintf("view-receipt?id=%v", common.EncodeID(receiptID))
		tgMessage := tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: tgChatID,
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
		tgApi := tgbotapi.NewBotAPIWithClient(toUserBotSettings.Token, urlfetch.Client(c))
		if _, err := tgApi.Send(tgMessage); err != nil {
			if _, forbidden := err.(tgbotapi.ErrAPIForbidden); forbidden {
				log.Infof(c, "Failed to send notification to user by Telegram: %v", err)
				editMessageConfig := tgbotapi.NewEditMessageText( // TODO: Add option buttons to send receipt
					transfer.Creator().TgChatID,
					int(transfer.CreatorTgReceiptByTgMsgID),
					"",
					emoji.ERROR_ICON+translator.Translate(trans.MESSAGE_TEXT_RECEIPT_NOT_SENT_AS_COUNTERPARTY_HAS_DISABLED_TG_BOT, locale.Code5, transfer.Counterparty().ContactName),
				)
				if _, err2 := tgApi.Send(editMessageConfig); err2 != nil {
					log.Errorf(c, "Failed to update creator's receipt message: %v", err2)
				}
				if err2 := gae_host.MarkTelegramChatAsForbidden(c, toUserBotSettings.Code, tgChatID, time.Now()); err2 != nil {
					log.Errorf(c, "Failed to call MarkTelegramChatAsStopped(): %v", err2.Error())
				}
			} else {
				log.Errorf(c, "Failed to send notification to user by Telegram: %v", err)
				editMessageConfig := tgbotapi.NewEditMessageText(
					transfer.Creator().TgChatID,
					int(transfer.CreatorTgReceiptByTgMsgID),
					"",
					emoji.ERROR_ICON+" "+err.Error(),
				)
				if _, err2 := tgApi.Send(editMessageConfig); err2 != nil {
					log.Errorf(c, "Failed to update creator's receipt message: %v", err2)
				}
			}
			if err2 := DelayOnReceiptSendFail(c, receiptID, time.Now(), err.Error()); err2 != nil {
				log.Errorf(c, errors.Wrap(err2, "Failed to update receipt entity with error info").Error())
			}
			return nil
		} else {
			log.Infof(c, "Notification sent to user by Telegram. Bot=%v, NotificationID=%v", toUserBotSettings.Code, receiptID)
			if err = DelayOnReceiptSentSuccess(c, time.Now(), receiptID, transferID, tgChatID, toUserBotSettings.Code); err != nil {
				log.Errorf(c, errors.Wrap(err, "Failed to call DelayOnReceiptSentSuccess()").Error())
			}
		}
	}

	{
		//var fromUserBotSettings bots.BotSettings
		//var ok bool
		//if transfer.Creator().TgBotID != "" {
		//	fromUserBotSettings, ok = telegram.BotsBy(c).ByCode[transfer.Creator().TgBotID] // TODO: This is wrong way to choose bot!
		//	if !ok {
		//		log.Errorf(c, "Bot settings not found for transfer(%v).Creator().TgBotID: [%v]", transferID, transfer.Creator().TgBotID)
		//	}
		//} else {
		//	log.Warningf(c, "Transfer.Creator().TgBotID is empty")
		//}
		//if !ok {
		//	fromUser, err := dal.User.GetUserByID(c, transfer.CreatorUserID)
		//	if err != nil {
		//		return err
		//	}
		//	localeCode := fromUser.PreferredLocale()
		//	fromUserBotSettings, err = telegram.GetBotSettingsByLang(c, localeCode, env)
		//	if err != nil {
		//		log.Warningf(c, "User has unknown locale: %v", localeCode)
		//		localeCode = strongo.LocaleEnUS.Code5
		//		fromUserBotSettings, err = telegram.GetBotSettingsByLang(c, localeCode, env)
		//		if err != nil {
		//			panic(errors.Wrap(err, "Bot settings bot found by locale").Error())
		//		}
		//	}
		//}
		//tgApi := tgbotapi.NewBotAPIWithClient(fromUserBotSettings.Token, urlfetch.Client(c))
		//
		//editMessageConfig := tgbotapi.NewEditMessageText( // TODO: Use creator's locale
		//	transfer.Creator().TgChatID,
		//	int(transfer.CreatorTgReceiptByTgMsgID),
		//	"",
		//	"\xF0\x9F\x93\xA4 "+
		//		translator.Translate(trans.MESSAGE_TEXT_RECEIPT_SENT_THROW_TELEGRAM, locale.Code5)+
		//		"\n\n"+ translator.Translate(trans.MESSAGE_TEXT_ASK_FOR_FEEDBAСK, locale.Code5),
		//)
		//keyboard := tgbotapi.NewInlineKeyboardMarkup(
		//	[]tgbotapi.InlineKeyboardButton{
		//		{Text: translator.Translate(trans.COMMAND_TEXT_GIVE_FEEDBACK, locale.Code5), CallbackData: "feedback"},
		//	},
		//)
		//editMessageConfig.ReplyMarkup = keyboard
		//if _, err = tgApi.Send(editMessageConfig); err != nil {
		//	if err.Error() == "Bad Request: message to edit not found" {
		//		log.Warningf(c, "Creator's receipt message is no longer exists: %v", err) // TODO: Handle gracefully or change to log.Info()
		//	} else {
		//		log.Errorf(c, "Failed to update creator's receipt message: %v", err)
		//		// TODO: Queue resending the message?
		//	}
		//	return nil
		//}
	}
	return err
})

func (_ UserDalGae) DelayUpdateUserHasDueTransfers(c context.Context, userID int64) error {
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
