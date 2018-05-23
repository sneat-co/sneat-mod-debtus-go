package dtb_transfer

import (
	"fmt"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/dtb_common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	//"bitbucket.org/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
	//"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	//"github.com/strongo/app/gaestandard"
	//"github.com/strongo/app"
	//"github.com/strongo/app/gae"
	//"context"
	//"google.golang.org/appengine/delay"
	//"google.golang.org/appengine/urlfetch"
	//"net/http"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var RemindAgainCallbackCommand = bots.NewCallbackCommand(dtb_common.CALLBACK_REMIND_AGAIN,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		q := callbackUrl.Query()
		var reminderID int64
		if reminderID, err = common.DecodeID(q.Get("id")); err != nil {
			return m, err
		} else if reminderID == 0 {
			return m, errors.New("reminderID == 0")
		}

		remindIn := q.Get("in")
		var remindInDuration time.Duration
		if remindIn == dtb_common.C_REMIND_IN_DISABLE {
			reportReminderIsActed(whc, "reminder-acted-disabled")
			// Do nothing? Empty duration means we need to disable reminder
		} else {
			if strings.HasSuffix(remindIn, "d") {
				// TODO: Temporary fix? Replaces 1d, 7d, 30d with hours
				if remindInDays, err := strconv.Atoi(remindIn[0 : len(remindIn)-1]); err == nil {
					remindIn = fmt.Sprintf("%vh", remindInDays*24)
				} else {
					log.Errorf(whc.Context(), errors.Wrap(err, "Failed to parse duration days").Error())
				}
			}
			if remindInDuration, err = time.ParseDuration(remindIn); err != nil {
				return m, err
			}
		}
		return rescheduleReminder(whc, reminderID, remindInDuration)
	},
)

func rescheduleReminder(whc bots.WebhookContext, reminderID int64, remindInDuration time.Duration) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	var oldReminder, newReminder models.Reminder

	if oldReminder, newReminder, err = dal.Reminder.RescheduleReminder(c, reminderID, remindInDuration); err != nil {
		if err == dal.ErrReminderAlreadyRescheduled {
			m = whc.NewMessageByCode(trans.MESSAGE_TEXT_REMINDER_ALREADY_RESCHEDULED)
			return m, nil
		}
		return
	}

	reportReminderIsActed(whc, "reminder-acted-rescheduled")

	if m.Text != "" {
		return m, err
	}
	var transfer models.Transfer
	if transfer, err = facade.GetTransferByID(c, oldReminder.TransferID); err != nil {
		return m, errors.Wrap(err, "Failed to get transferEntity by id")
	}
	var messageText string
	if remindInDuration == time.Duration(0) {
		messageText = whc.Translate(trans.MESSAGE_TEXT_REMINDER_DISABLED)
	} else {
		messageText = whc.Translate(trans.MESSAGE_TEXT_REMINDER_SET, newReminder.DtNext.Format("Mon, 2 Jan 15:04:05 MST (-0700) 2006"))
	}

	chatEntity := whc.ChatEntity()
	if chatEntity.IsAwaitingReplyTo(SET_NEXT_REMINDER_DATE_COMMAND) {
		chatEntity.SetAwaitingReplyTo("")
	}

	if m, err = dtb_general.EditReminderMessage(whc, transfer, messageText); err != nil {
		return
	}

	if remindInDuration == time.Duration(0) {
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{
					Text: whc.Translate(trans.COMMAND_TEXT_REMINDER_ENABLE),
					CallbackData: fmt.Sprintf("%v?reminder=%v&transfer=%v", ENABLE_REMINDER_AGAIN_COMMAND,
						common.EncodeID(reminderID), common.EncodeID(transfer.ID),
					),
				},
			},
		)
	}

	//go func() {
	//	chatID := whc.MustBotChatID()
	//	intChatID, err := strconv.ParseInt(chatID, 10, 64)
	//	if err != nil {
	//		log.Errorf(c, "Failed to parse BotChatID to int: %v\nwhc.BotChatID(): %v", err, chatID)
	//		return
	//	}
	//	if err = delayAskForFeedback(c, whc.GetBotCode(), intChatID, whc.AppUserIntID()); err != nil {
	//		log.Errorf(c, "Failed to create task for asking feedback: %v", err)
	//	}
	//}()

	return
}

//const ASK_FOR_FEEDBACK_TASK = "ask-for-feedback"
//
//func delayAskForFeedback(c context.Context, botCode string, chatID int64, userID int64) error {
//	task, err := gae.CreateDelayTask(common.QUEUE_CHATS, ASK_FOR_FEEDBACK_TASK, delayedAskForFeedback, botCode, chatID, userID)
//	if err != nil {
//		return err
//	}
//	task.Delay = time.Second / 2
//	task, err = gae.AddTaskToQueue(c, task, common.QUEUE_CHATS)
//	return err
//}
//
//var delayedAskForFeedback = delay.Func(ASK_FOR_FEEDBACK_TASK,
//	func(c context.Context, botID string, chatID, userID int64) error {
//		log.Debugf(c, "delayedAskForFeedback(botID=%v, chatID=%d, userID=%d)", botID, chatID, userID)
//		if botSettings, ok := telegram.Bots(gaestandard.GetEnvironment(c), nil).ByCode[botID]; !ok {
//			log.Errorf(c, "Bot settings not found by ID: "+botID)
//			return nil
//		} else {
//			locale, err := facade.GetLocale(c, botID, chatID, userID)
//			if err != nil {
//				return err
//			}
//			translator := strongo.NewSingleMapTranslator(locale, strongo.NewMapTranslator(c, trans.TRANS))
//			text := translator.Translate(trans.MESSAGE_TEXT_ASK_FOR_FEEDBAСK)
//			messageConfig := tgbotapi.NewMessage(chatID, text)
//			messageConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
//				[]tgbotapi.InlineKeyboardButton{
//					{Text: translator.Translate(trans.COMMAND_TEXT_GIVE_FEEDBACK), CallbackData: "feedback"},
//				},
//			)
//			messageConfig.ParseMode = "HTML"
//			tgBotApi := tgbotapi.NewBotAPIWithClient(botSettings.Token, &http.Client{Transport: &urlfetch.Transport{Context: c}})
//			if message, err := tgBotApi.Send(messageConfig); err != nil {
//				log.Debugf(c, "Faield to send message to Telegram: %v", err)
//				return nil
//			} else {
//				log.Debugf(c, "Sent to Telegram: %v", message.MessageID)
//			}
//		}
//		return nil
//	})

//func disableReminders(whc bots.WebhookContext, transferID int64) (m bots.MessageFromBot, err error) {
//	c := whc.Context()
//	transferKey, transfer, err := facade.GetTransferByID(c, transferID)
//	userID := whc.AppUserIntID()
//	if !transfer.IsRemindersDisabled(userID) {
//		err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
//			transferKey, transfer, err = gaedal.GetTransferByID(tc, transferID)
//			if err != nil {
//				return err
//			}
//			changed := false
//			if !transfer.IsRemindersDisabled(userID) {
//				transfer.DisableAutoReminders(userID)
//				changed = true
//			}
//			if transfer.IsDue2Notify {
//				isDue2Notify := !transfer.IsRemindersDisabled(transfer.CreatorUserID) || !transfer.IsRemindersDisabled(transfer.CreatorCounterparty().UserID)
//				if !isDue2Notify {
//					transfer.IsDue2Notify = false
//					changed = true
//				}
//			}
//			if changed {
//				_, err = nds.Put(tc, transferKey, transfer)
//			}
//			return err
//		}, nil)
//		if err != nil {
//			return m, err
//		}
//	}
//	m = dtb_general.EditReminderMessage(whc, transferID, transfer, whc.Translate(trans.MESSAGE_TEXT_REMINDER_DISABLED))
//	return m, err
//}
