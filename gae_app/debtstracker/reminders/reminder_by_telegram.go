package reminders

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/dtb_common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/analytics"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/gaestandard"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/delay"
	"google.golang.org/appengine/v2/urlfetch"
	"net/http"
	"time"
)

func sendReminderByTelegram(c context.Context, transfer models.Transfer, reminder models.Reminder, tgChatID int64, tgBot string) (sent, channelDisabledByUser bool, err error) {
	log.Debugf(c, "sendReminderByTelegram(transfer.ID=%v, reminder.ID=%v, tgChatID=%v, tgBot=%v)", transfer.ID, reminder.ID, tgChatID, tgBot)

	if tgChatID == 0 {
		panic("tgChatID == 0")
	}
	if tgBot == "" {
		panic("tgBot is empty string")
	}

	var locale strongo.Locale

	if locale, err = facade.GetLocale(c, tgBot, tgChatID, reminder.UserID); err != nil {
		return
	}

	//if !tgChat.DtForbidden.IsZero() {
	//	log.Infof(c, "Telegram chat(id=%v) is not available since: %v", tgChatID, tgChat.DtForbidden)
	//	return false
	//}

	translator := strongo.NewSingleMapTranslator(locale, strongo.NewMapTranslator(c, trans.TRANS))

	env := gaestandard.GetEnvironment(c)

	if botSettings, ok := tgbots.Bots(env, nil).ByCode[tgBot]; !ok {
		err = fmt.Errorf("bot settings not found (env=%v, tgBotID=%v)", env, tgBot)
		return
	} else {
		tgBotApi := tgbotapi.NewBotAPIWithClient(botSettings.Token, &http.Client{Transport: &urlfetch.Transport{Context: c}})
		messageText := fmt.Sprintf(
			"<b>%v</b>\n%v\n\n",
			translator.Translate(trans.MESSAGE_TEXT_REMINDER),
			translator.Translate(trans.MESSAGE_TEXT_REMINDER_ASK_IF_RETURNED),
		)

		executionContext := strongo.NewExecutionContext(c, translator)
		utm := common.UtmParams{
			Source:   "TODO",
			Medium:   telegram.PlatformID,
			Campaign: common.UTM_CAMPAIGN_REMINDER,
		}
		messageText += common.TextReceiptForTransfer(executionContext, transfer, reminder.UserID, common.ShowReceiptToAutodetect, utm)

		messageConfig := tgbotapi.NewMessage(tgChatID, messageText)

		var db dal.Database
		if db, err = facade.GetDatabase(c); err != nil {
			return
		}
		err = db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) (err error) {
			reminder, err = dtdal.Reminder.GetReminderByID(c, tx, reminder.ID)
			if err != nil {
				return err
			}
			callbackData := fmt.Sprintf(dtb_common.DEBT_RETURN_CALLBACK_DATA, dtb_common.CALLBACK_DEBT_RETURNED_PATH, common.EncodeIntID(reminder.ID), "%v")
			messageConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{
					{Text: translator.Translate(trans.COMMAND_TEXT_REMINDER_RETURNED_IN_FULL), CallbackData: fmt.Sprintf(callbackData, dtb_common.RETURNED_FULLY)},
				},
				[]tgbotapi.InlineKeyboardButton{
					{Text: translator.Translate(trans.COMMAND_TEXT_REMINDER_RETURNED_PARTIALLY), CallbackData: fmt.Sprintf(callbackData, dtb_common.RETURNED_PARTIALLY)},
				},
				[]tgbotapi.InlineKeyboardButton{
					{Text: translator.Translate(trans.COMMAND_TEXT_REMINDER_NOT_RETURNED), CallbackData: fmt.Sprintf(callbackData, dtb_common.RETURNED_NOTHING)},
				},
			)
			messageConfig.ParseMode = "HTML"
			message, err := tgBotApi.Send(messageConfig)
			if err != nil {
				if _, isForbidden := err.(tgbotapi.ErrAPIForbidden); isForbidden { // TODO: Mark chat as deleted?
					log.Infof(c, "Telegram bot API returned status 'forbidden' - either issue with token or chat deleted by user")
					if err2 := DelaySetChatIsForbidden(c, botSettings.Code, tgChatID, time.Now()); err2 != nil {
						log.Errorf(c, "Failed to delay to set chat as forbidden: %v", err2)
					}
					channelDisabledByUser = true
					return nil // Do not pass error up
				} else {
					log.Debugf(c, "messageConfig.Text: %v", messageConfig.Text)
					return fmt.Errorf("Failed in call to Telegram API: %w", err)
				}
			}
			sent = true
			log.Infof(c, "Sent message to telegram. MessageID: %v", message.MessageID)

			if err = dtdal.Reminder.SetReminderIsSentInTransaction(tc, tx, reminder, time.Now(), int64(message.MessageID), "", locale.Code5, ""); err != nil {
				err = dtdal.Reminder.DelaySetReminderIsSent(tc, reminder.ID, time.Now(), int64(message.MessageID), "", locale.Code5, "")
			}
			//
			return
		}, nil)

		if err != nil {
			log.Errorf(c, fmt.Errorf("error while sending by Telegram: %w", err).Error())
			return
		}
		if sent {
			analytics.ReminderSent(c, reminder.UserID, translator.Locale().Code5, telegram.PlatformID)
		}
	}
	return
}

func DelaySetChatIsForbidden(c context.Context, botID string, tgChatID int64, at time.Time) error {
	return gae.CallDelayFunc(c, common.QUEUE_CHATS, "set-chat-is-forbidden", delaySetChatIsForbidden, botID, tgChatID, at)
}

var delaySetChatIsForbidden = delay.MustRegister("SetChatIsForbidden", SetChatIsForbidden)

func SetChatIsForbidden(c context.Context, botID string, tgChatID int64, at time.Time) error {
	log.Debugf(c, "SetChatIsForbidden(tgChatID=%v, at=%v)", tgChatID, at)
	panic("TODO: Implement SetChatIsForbidden")
	//err := gaehost.MarkTelegramChatAsForbidden(c, botID, tgChatID, at)
	//if err == nil {
	//	log.Infof(c, "Success")
	//} else {
	//	log.Errorf(c, err.Error())
	//	if err == datastore.ErrNoSuchEntity {
	//		return nil // Do not re-try
	//	}
	//}
	//return err
}
