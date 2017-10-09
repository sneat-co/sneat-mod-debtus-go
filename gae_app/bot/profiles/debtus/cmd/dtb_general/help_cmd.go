package dtb_general

import (
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

const HELP_COMMAND = "help"

var HelpCommand = bots.Command{
	Code:     HELP_COMMAND,
	Icon:     emoji.HELP_ICON,
	Commands: trans.Commands(trans.COMMAND_HELP),
	Title:    trans.COMMAND_TEXT_HELP,
	Titles:   map[string]string{bots.SHORT_TITLE: ""},
	Action: func(whc bots.WebhookContext) (bots.MessageFromBot, error) {
		return helpCommandAction(whc, true)
	},
}

func helpCommandAction(whc bots.WebhookContext, showFeedbackButton bool) (m bots.MessageFromBot, err error) {
	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text: emoji.PUBLIC_LOUDSPEAKER + " " + whc.Translate(trans.COMMAND_TEXT_OPEN_USER_REPORT),
				URL:  getUserReportUrl(whc, ""),
			},
		},
		[]tgbotapi.InlineKeyboardButton{btnSubmitBug(whc, getUserReportUrl(whc, "bug"))},
		[]tgbotapi.InlineKeyboardButton{btnSubmitIdea(whc, getUserReportUrl(whc, "idea"))},
	)
	if showFeedbackButton {
		keyboardMarkup.InlineKeyboard = append(
			keyboardMarkup.InlineKeyboard,
			[]tgbotapi.InlineKeyboardButton{
				{
					Text:         emoji.STAR_ICON + " " + whc.Translate(trans.COMMAND_TEXT_ASK_FOR_FEEDBACK),
					CallbackData: FEEDBACK_COMMAND,
				},
			})
	}
	if showFeedbackButton {
		m = whc.NewMessageByCode(trans.MESSAGE_TEXT_HELP)
		m.Keyboard = keyboardMarkup
	} else {
		if m, err = whc.NewEditMessage("", bots.MessageFormatText); err != nil {
			return
		}
		m.Keyboard = keyboardMarkup
	}

	return m, err
}

func getUserReportUrl(t strongo.SingleLocaleTranslator, submit string) string {
	switch t.Locale().Code5 {
	case strongo.LOCALE_RU_RU:
		switch submit {
		case "idea":
			return "https://goo.gl/dAKHFC"
		case "bug":
			return "https://goo.gl/jQM2K5"
		case "":
			return "https://goo.gl/Vge31X"
		default:
			panic("Parameter 'submit' should be either 'idea' or 'bug'")
		}
	default:
		switch submit {
		case "idea":
			return "https://goo.gl/sl09Wr"
		case "bug":
			return "https://goo.gl/x5H6Fn"
		case "":
			return "https://goo.gl/3tB0FG"
		default:
			panic("Parameter 'submit' should be either 'idea' or 'bug'")
		}
	}
}

func btnSubmitIdea(whc bots.WebhookContext, url string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.InlineKeyboardButton{
		Text: emoji.BULB_ICON + " " + whc.Translate(trans.COMMAND_TEXT_SUBMIT_AN_IDEA),
		URL:  url,
	}
}

func btnSubmitBug(whc bots.WebhookContext, url string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.InlineKeyboardButton{
		Text: emoji.ERROR_ICON + " " + whc.Translate(trans.COMMAND_TEXT_REPORT_A_BUG),
		URL:  url,
	}
}

const ADS_COMMAND = "ads"

var AdsCommand = bots.Command{
	Code:     ADS_COMMAND,
	Icon:     emoji.NEWSPAPER_ICON,
	Commands: []string{emoji.NEWSPAPER_ICON, "/ads", "/реклама"},
	Title:    trans.COMMAND_TEXT_HELP,
	Titles:   map[string]string{bots.SHORT_TITLE: ""},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		chatEntity := whc.ChatEntity()

		yesOption := emoji.PHONE_ICON + " " + whc.Translate(trans.COMMAND_TEXT_SUBSCRIBE_TO_APP)
		noOption := whc.Translate(trans.COMMAND_TEXT_I_AM_FINE_WITH_BOT)
		if chatEntity.GetAwaitingReplyTo() == "" {
			chatEntity.SetAwaitingReplyTo(ADS_COMMAND)
			m = whc.NewMessage(emoji.NEWSPAPER_ICON + " " + whc.Translate(trans.MESSAGE_TEXT_YOUR_ABOUT_ADS))
			m.DisableWebPagePreview = true
			m.Keyboard = tgbotapi.NewReplyKeyboard(
				[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(yesOption)},
				[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(noOption)},
				[]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(MainMenuCommand.DefaultTitle(whc))},
			)
		} else {
			switch whc.Input().(bots.WebhookTextMessage).Text() {
			case yesOption:
				m = whc.NewMessageByCode(trans.MESSAGE_TEXT_SUBSCRIBED_TO_APP)
				SetMainMenuKeyboard(whc, &m)
				chatEntity.SetAwaitingReplyTo("")
			case noOption:
				m = whc.NewMessageByCode(trans.MESSAGE_TEXT_NOT_INTERESTED_IN_APP)
				SetMainMenuKeyboard(whc, &m)
				chatEntity.SetAwaitingReplyTo("")
			default:
				m = whc.NewMessageByCode(trans.MESSAGE_TEXT_PLEASE_CHOOSE_FROM_OPTIONS_PROVIDED)
			}
		}
		return m, err
	},
}
