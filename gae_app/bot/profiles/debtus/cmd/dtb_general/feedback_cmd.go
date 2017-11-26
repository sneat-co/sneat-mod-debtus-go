package dtb_general

import (
	"fmt"
	"net/url"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	//"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/dtb_common"
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/admin"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
)

const (
	FEEDBACK_COMMAND   = "feedback"
	FEEDBACK_UNDECIDED = "undecided"
)

func AskToTranslate(t strongo.SingleLocaleTranslator) string {
	return strings.Replace(t.Translate(trans.MESSAGE_TEXT_ASK_TO_TRANSLATE),
		"<a>",
		`<a href="https://goo.gl/tZsqW1">`, // https://github.com/DebtsTracker/translations
		1)
}

func YouCanHelp(t strongo.SingleLocaleTranslator, s, botCode string) string {
	s = t.Translate(s)
	s = strings.Replace(s, "<a storebot>", Ahref(StorebotUrl(botCode)), 1)
	s = strings.Replace(s, "<a share-vk>", Ahref(ShareToVkUrl()), 1)
	s = strings.Replace(s, "<a share-fb>", Ahref(ShareToFacebookUrl()), 1)
	s = strings.Replace(s, "<a share-twitter>", Ahref(ShareToTwitter()), 1)
	return s
}

func FeedbackLinks(t strongo.SingleLocaleTranslator, s string) string {
	s = strings.Replace(s, "<a suggest-idea>", Ahref(getUserReportUrl(t, "idea")), 1)
	s = strings.Replace(s, "<a submit-bug>", Ahref(getUserReportUrl(t, "bug")), 1)
	return s
}

func Ahref(url string) string {
	return fmt.Sprintf(`<a href="%v">`, url)
}

func StorebotUrl(botID string) string {
	return "https://t.me/storebot?start=" + botID
}

func ShareToFacebookUrl() string {
	return "https://goo.gl/WyrRLg" // "https://www.facebook.com/sharer/sharer.php?u=https%3A//debtstracker.io/"
}

func ShareToVkUrl() string {
	return "https://goo.gl/lcnPJ3" // "https://vk.com/share.php?url=https%3A//debtstracker.io/&title=Отличный%20Telegram%20бот%20для%20учёта%20долгов%20-%20https%3A//t.me/DebtsTrackerRuBot"
}

func ShareToTwitter() string {
	return "https://goo.gl/Xbv004" // "https://twitter.com/home?status=The%20%40DebtsTracker%20is%20awesome.%20Check%20their%20%23Telegram%20bot%20https%3A//t.me/DebtsTrackerBot"
}

/*
var FeedbackCallbackCommand = bots.NewCallbackCommand(FEEDBACK_COMMAND, func(whc bots.WebhookContext, callbackUrl *url.URL) (bots.MessageFromBot, error) {
	return FeedbackCommand.Action(whc)
})

var FeedbackCommand = bots.Command{
	Code:     FEEDBACK_COMMAND,
	Commands: trans.Commands(trans.COMMAND_TEXT_FEEDBACK),
	Title:    trans.COMMAND_TEXT_HIGH_FIVE,
	Icon:     emoji.STAR_ICON,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		FeedbackCommand.Action(whc)
		chatEntity := whc.ChatEntity()
		switch chatEntity.GetAwaitingReplyTo() {
		//case "":
		//	return showFeedbackOptions(whc, chatEntity)
		case FEEDBACK_COMMAND:
			mt := whc.Input().(bots.WebhookTextMessage).Text()
			words := strings.SplitN(mt, " ", 2)
			feedbackEntity := models.FeedbackEntity{
				UserID: whc.AppUserIntID(),
			}
			//mainMenuButton := []tgbotapi.InlineKeyboardButton{
			//	{
			//		Text: whc.CommandText(trans.COMMAND_TEXT_MAIN_MENU_TITLE, emoji.MAIN_MENU_ICON),
			//		CallbackData: MAIN_MENU_COMMAND,
			//	},
			//}

			switch words[0] {
			case emoji.EMO_SMILING_ICON:
				feedbackEntity.Rate = "Positive"
				thankYouText := strings.Replace(
					whc.Translate(trans.MESSAGE_TEXT_ON_FEEDBACK_POSITIVE),
					fmt.Sprintf("{{%v}}", trans.MESSAGE_TEXT_YOU_CAN_HELP_BY),
					YouCanHelp(whc, trans.MESSAGE_TEXT_YOU_CAN_HELP_BY, whc.GetBotCode()),
					1)
				thankYouText = FeedbackLinks(whc, thankYouText)
				m = whc.NewMessage(emoji.EMO_SMILING_RED_CHEEKS + " " + thankYouText + "\n" + AskToTranslate(whc))
			case emoji.EMO_NEUTRAL:
				feedbackEntity.Rate = "Neutral"
				text := FeedbackLinks(whc, whc.Translate(trans.MESSAGE_TEXT_ON_FEEDBACK_NEUTRAL))
				m = whc.NewMessage(emoji.EMO_CONFUSED + " " + text + "\n\n" + AskToTranslate(whc))
				m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
					[]tgbotapi.InlineKeyboardButton{btnSubmitIdea(whc, getUserReportUrl(whc, "idea"))},
					[]tgbotapi.InlineKeyboardButton{btnSubmitBug(whc, getUserReportUrl(whc, "bug"))},
					//mainMenuButton,
				)
			case emoji.EMO_ANGRY_ICON:
				feedbackEntity.Rate = "Angry"
				text := FeedbackLinks(whc, whc.Translate(trans.MESSAGE_TEXT_ON_FEEDBACK_NEGATIVE))
				m = whc.NewMessage(emoji.EMO_EMBARRASSED + " " + text)
				m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
					[]tgbotapi.InlineKeyboardButton{btnSubmitBug(whc, getUserReportUrl(whc, "bug"))},
					[]tgbotapi.InlineKeyboardButton{btnSubmitIdea(whc, getUserReportUrl(whc, "idea"))},
					//mainMenuButton,
				)
			case emoji.EMO_THINKING:
				feedbackEntity.Rate = FEEDBACK_UNDECIDED
			default:
				m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_PLEASE_CHOOSE_FROM_OPTIONS_PROVIDED))
				m.Keyboard = feedbackOptionsTelegramKeyboard(whc)
				return m, nil
			}
			m.DisableWebPagePreview = true

			c := whc.Context()
			whc.GetAppUser()
			if _, _, err = facade.SaveFeedback(c, &feedbackEntity); err != nil {
				return m, errors.Wrap(err, "Failed to save Feedback to DB")
			}
			if feedbackEntity.Rate == FEEDBACK_UNDECIDED {
				return MainMenuAction(whc, "", false)
			} else {
				//if _, err = whc.Responder().SendMessage(c, m, bots.BotApiSendMessageOverHTTPS); err != nil {
				//	return m, err
				//}
				//m = whc.NewMessageByCode(trans.MESSAGE_TEXT_BACK_TO_MAIN_MENU)
				m.Keyboard = tgbotapi.NewReplyKeyboard(
					[]tgbotapi.KeyboardButton{{Text: whc.CommandText(trans.COMMAND_TEXT_MAIN_MENU_TITLE, emoji.MAIN_MENU_ICON)}},
				)
				return m, err
			}
		default:
			return showFeedbackOptions(whc, chatEntity)
		}
	},
	CallbackAction: func(whc bots.WebhookContext, _ *url.URL) (m bots.MessageFromBot, err error) {
		m, err = showFeedbackOptions(whc, whc.ChatEntity())
		if _, err = whc.Responder().SendMessage(whc.Context(), m, bots.BotApiSendMessageOverHTTPS); err != nil {
			return m, err
		}
		return helpCommandAction(whc, false)
	},
}
*/

func feedbackCommandAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	m = whc.NewMessageByCode(trans.MESSAGE_TEXT_DO_YOU_LIKE_OUR_BOT)
	m.Text = strings.Replace(m.Text, "{{bot}}", whc.GetBotCode(), 1)
	m.Keyboard = feedbackOptionsTelegramKeyboard(whc)
	return m, err
}

var FeedbackCommand = bots.Command{
	Code:     FEEDBACK_COMMAND,
	Title:    trans.COMMAND_TEXT_FEEDBACK,
	Commands: trans.Commands(trans.COMMAND_TEXT_FEEDBACK, FEEDBACK_COMMAND, emoji.STAR_ICON),
	Icon:     emoji.STAR_ICON,
	Action:   feedbackCommandAction,
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		like := callbackUrl.Query().Get("like")
		if like == "" {
			m, err = feedbackCommandAction(whc)
			return
		}
		feedbackEntity := models.FeedbackEntity{
			UserID: whc.AppUserIntID(),
			CreatedOn: general.CreatedOn{
				CreatedOnPlatform: whc.BotPlatform().Id(),
				CreatedOnID:       whc.GetBotCode(),
			},
		}
		switch like {
		case "yes":
			feedbackEntity.Rate = "like"
		case "no":
			feedbackEntity.Rate = "dislike"
		default:
			err = fmt.Errorf("Unexpected 'like' value: %v", like)
			return
		}
		var feedback models.Feedback
		if err = dal.DB.RunInTransaction(whc.Context(), func(c context.Context) (err error) {
			if feedback, _, err = facade.SaveFeedback(c, 0, &feedbackEntity); err != nil {
				return
			}
			return nil
		}, dal.CrossGroupTransaction); err != nil {
			return
		}
		switch like {
		case "yes":
			m, err = askIfCanRateAtStoreBot(whc)
		case "no":
			m, err = askToWriteFeedback(whc, feedback.ID)
		}
		return
	},
}

func feedbackOptionsTelegramKeyboard(whc bots.WebhookContext) *tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_YES_EXCLAMATION, emoji.GREEN_CHECKBOX), CallbackData: FEEDBACK_COMMAND + "?like=yes"},
			{Text: whc.Translate(trans.COMMAND_TEXT_NOT_TOO_MUCH, emoji.CROSS_MARK), CallbackData: FEEDBACK_COMMAND + "?like=no"},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_WRITE_FEEDBACK, emoji.MEMO_ICON), CallbackData: FEEDBACK_TEXT_COMMAND},
		},
	)
}

func askIfCanRateAtStoreBot(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	m, err = editTelegramMessageText(whc, "", whc.Translate(trans.MESSAGE_TEXT_CAN_YOU_RATE_AT_STOREBOT))
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_YES, emoji.GREEN_CHECKBOX), CallbackData: CAN_YOU_RATE_COMMAND + "?will-rate=yes"},
			{Text: whc.Translate(trans.COMMAND_TEXT_NO, emoji.CROSS_MARK), CallbackData: CAN_YOU_RATE_COMMAND + "?will-rate=no"},
		},
	)
	return
}

const CAN_YOU_RATE_COMMAND = "can-you-rate"

var CanYouRateCommand = bots.Command{
	Code: CAN_YOU_RATE_COMMAND,
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		log.Debugf(whc.Context(), "CanYouRateCommand.CallbackAction): whc.ChatEntity().GetPreferredLanguage()=%v", whc.ChatEntity().GetPreferredLanguage())
		if callbackUrl == nil || callbackUrl.RawQuery == "" {
			m, err = askIfCanRateAtStoreBot(whc)
		} else {
			switch callbackUrl.Query().Get("will-rate") {
			case "yes":
				m, err = editTelegramMessageText(whc, "", strings.Replace(whc.Translate(trans.MESSAGE_TEXT_HOW_TO_RATE_AT_STOREBOT), "{{bot}}", whc.GetBotCode(), 1))
			case "no":
				thankYouText := strings.Replace(
					whc.Translate(trans.MESSAGE_TEXT_ON_REFUSED_TO_RATE),
					fmt.Sprintf("{{%v}}", trans.MESSAGE_TEXT_YOU_CAN_HELP_BY),
					YouCanHelp(whc, trans.MESSAGE_TEXT_YOU_CAN_HELP_BY, whc.GetBotCode()),
					1)
				thankYouText = FeedbackLinks(whc, thankYouText)
				if m, err = editTelegramMessageText(whc, "/", thankYouText); err != nil {
					return
				}
				m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
					[]tgbotapi.InlineKeyboardButton{
						{Text: whc.Translate(trans.COMMAND_TEXT_WRITE_FEEDBACK, emoji.MEMO_ICON), CallbackData: FEEDBACK_TEXT_COMMAND},
					},
					[]tgbotapi.InlineKeyboardButton{
						{Text: emoji.MAIN_MENU_ICON + " " + whc.Translate(trans.COMMAND_TEXT_MAIN_MENU_TITLE), CallbackData: MAIN_MENU_COMMAND},
					},
				)
			default:
				m = whc.NewMessage(fmt.Sprintf("Unknown 'will-rate' value, expected yes/no, got: %v", callbackUrl.Query().Get("reply")))
				log.Errorf(whc.Context(), m.Text)
			}
		}
		return
	},
}

func askToWriteFeedback(whc bots.WebhookContext, feedbackID int64) (m bots.MessageFromBot, err error) {
	m = whc.NewMessageByCode(trans.MESSAGE_TEXT_ASK_TO_WRITE_FEEDBACK_WITHIN_MESSENGER)
	//m, err = editTelegramMessageText(whc, FEEDBACK_TEXT_COMMAND, whc.Translate(trans.MESSAGE_TEXT_ASK_TO_WRITE_FEEDBACK_WITHIN_MESSENGER))
	whc.ChatEntity().SetAwaitingReplyTo(FEEDBACK_TEXT_COMMAND)
	if feedbackID != 0 {
		whc.ChatEntity().AddWizardParam("feedback", strconv.FormatInt(feedbackID, 10))
	}
	m.Keyboard = tgbotapi.NewHideKeyboard(false)
	return
}

func editTelegramMessageText(whc bots.WebhookContext, awaitingReplyTo, text string) (m bots.MessageFromBot, err error) {
	var (
		tgChatID int64
		chatID   string
	)

	if chatID, err = whc.BotChatID(); err != nil {
		return
	}

	if tgChatID, err = strconv.ParseInt(chatID, 10, 64); err != nil {
		return
	}
	// TODO: Does it changes locale from RU to EN?
	messageID := whc.Input().(telegram_bot.TelegramWebhookCallbackQuery).GetMessage().IntID()
	if m, err = whc.NewEditMessage(text, bots.MessageFormatHTML); err != nil {
		return
	}
	m.EditMessageUID = telegram_bot.NewChatMessageUID(tgChatID, int(messageID))
	if awaitingReplyTo != "" {
		if awaitingReplyTo == "/" {
			awaitingReplyTo = ""
		}
		whc.ChatEntity().SetAwaitingReplyTo(awaitingReplyTo)
	}
	return
}

const FEEDBACK_TEXT_COMMAND = "feedback-text"

var FeedbackTextCommand = bots.Command{
	Code: FEEDBACK_TEXT_COMMAND,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		switch whc.Input().(type) {
		case bots.WebhookTextMessage:
			mt := whc.Input().(bots.WebhookTextMessage).Text()
			feedbackParam := whc.ChatEntity().GetWizardParam("feedback")

			var feedback models.Feedback
			c := whc.Context()
			if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
				if feedbackParam == "" {
					feedback.FeedbackEntity = &models.FeedbackEntity{
						Rate:   "none",
						UserID: whc.AppUserIntID(),
						Text:   mt,
						CreatedOn: general.CreatedOn{
							CreatedOnPlatform: whc.BotPlatform().Id(),
							CreatedOnID:       whc.GetBotCode(),
						},
					}
				} else {
					if feedback.ID, err = strconv.ParseInt(feedbackParam, 10, 64); err != nil {
						return
					}
					if feedback, err = dal.Feedback.GetFeedbackByID(c, feedback.ID); err != nil {
						return
					}
					feedback.Text = mt
				}
				if feedback, _, err = facade.SaveFeedback(c, 0, feedback.FeedbackEntity); err != nil {
					return
				}
				return nil
			}, dal.CrossGroupTransaction); err != nil {
				return
			}
			m = whc.NewMessageByCode(trans.MESSAGE_TEXT_THANKS)
			m.Text += fmt.Sprintf(` Feedback #<a href="https://debtstracker.io/app/#/feedback/%d">%d</a>`, feedback.ID, feedback.ID)
			SetMainMenuKeyboard(whc, &m)
			if err2 := admin.SendFeedbackToAdmins(c, feedback); err2 != nil {
				log.Errorf(c, errors.Wrap(err2, "Faield to notify admins").Error())
			}
		default:
			m = whc.NewMessageByCode(trans.MESSAGE_TEXT_PLEASE_SEND_TEXT)
		}
		return
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return askToWriteFeedback(whc, 0)
	},
}
