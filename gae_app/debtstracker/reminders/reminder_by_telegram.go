package reminders

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/analytics"
	"bitbucket.com/debtstracker/gae_app/bot/profiles/debtus/dtb_common"
	"bitbucket.com/debtstracker/gae_app/bot/platforms/telegram"
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"github.com/strongo/app/gae"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"time"
	"bitbucket.com/debtstracker/gae_app/debtstracker/facade"
	"bitbucket.com/debtstracker/gae_app/gaestandard"
)

func sendReminderByTelegram(c context.Context, transferReminderTo TransferReminderTo, transfer models.Transfer, reminderID, userID, tgChatID int64, tgBot string) (sent, channelDisabledByUser bool, err error) {
	log.Debugf(c, "sendReminderByTelegram(transferReminderTo:%v, transfer.ID=%d, reminderID=%d)", transferReminderTo, transfer.ID, reminderID)

	var locale strongo.Locale

	if locale, err = facade.GetLocale(c, tgBot, tgChatID, userID); err != nil {
		return
	}

	//if !tgChat.DtForbidden.IsZero() {
	//	log.Infof(c, "Telegram chat(id=%v) is not available since: %v", tgChatID, tgChat.DtForbidden)
	//	return false
	//}

	translator := strongo.NewSingleMapTranslator(locale, strongo.NewMapTranslator(c, trans.TRANS))

	if botSettings, ok := telegram.Bots(gaestandard.GetEnvironment(c), nil).ByCode[transfer.CreatorTgBotID]; ok {
		tgBotApi := tgbotapi.NewBotAPIWithClient(botSettings.Token, &http.Client{Transport: &urlfetch.Transport{Context: c}})
		messageText := fmt.Sprintf(
			"<b>%v</b>\n%v\n\n",
			translator.Translate(trans.MESSAGE_TEXT_REMINDER),
			translator.Translate(trans.MESSAGE_TEXT_REMINDER_ASK_IF_RETURNED),
		)

		executionContext := strongo.NewExecutionContext(c, translator)
		utm := common.UtmParams{
			Source:   "TODO",
			Medium:   telegram_bot.TelegramPlatformID,
			Campaign: common.UTM_CAMPAIGN_REMINDER,
		}
		messageText += common.TextReceiptForTransfer(executionContext, transfer, userID, common.ShowReceiptToAutodetect, utm)

		messageConfig := tgbotapi.NewMessage(tgChatID, messageText)

		err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
			reminder, err := dal.Reminder.GetReminderByID(c, reminderID)
			if err != nil {
				return err
			}
			callbackData := fmt.Sprintf(dtb_common.DEBT_RETURN_CALLBACK_DATA, dtb_common.CALLBACK_DEBT_RETURNED_PATH, common.EncodeID(reminderID), "%v")
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
					return errors.Wrap(err, "Failed in call to Telegram API")
				}
			}
			sent = true
			log.Infof(c, "Sent message to telegram. MessageID: %v", message.MessageID)

			if err = dal.Reminder.SetReminderIsSentInTransaction(tc, reminder, time.Now(), int64(message.MessageID), "", locale.Code5, ""); err != nil {
				err = dal.Reminder.DelaySetReminderIsSent(tc, reminderID, time.Now(), int64(message.MessageID), "", locale.Code5, "")
			}
			//
			return err
		}, nil)

		if sent {
			analytics.ReminderSent(c, userID, translator.Locale().Code5, telegram_bot.TelegramPlatformID)
		}

		if err != nil {
			log.Errorf(c, errors.Wrapf(err, "Error while sending by Telegram").Error())
		}
	}
	return
}

func DelaySetChatIsForbidden(c context.Context, botID string, tgChatID int64, at time.Time) error {
	return gae.CallDelayFunc(c, common.QUEUE_CHATS, "set-chat-is-forbidden", delaySetChatIsForbidden, botID, tgChatID, at)
}

var delaySetChatIsForbidden = delay.Func("SetChatIsForbidden", SetChatIsForbidden)

func SetChatIsForbidden(c context.Context, botID string, tgChatID int64, at time.Time) error {
	log.Debugf(c, "SetChatIsForbidden(tgChatID=%v, at=%v)", tgChatID, at)
	err := gae_host.MarkTelegramChatAsForbidden(c, botID, tgChatID, at)
	if err == nil {
		log.Infof(c, "Success")
	} else {
		log.Errorf(c, err.Error())
		if err == datastore.ErrNoSuchEntity {
			return nil // Do not re-try
		}
	}
	return err
}
