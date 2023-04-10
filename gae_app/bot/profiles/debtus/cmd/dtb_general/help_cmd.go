package dtb_general

import (
	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/emoji"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/app"
	bots "github.com/strongo/bots-framework/core"
)

func HelpCommandAction(whc botsfw.WebhookContext, showFeedbackButton bool) (m botsfw.MessageFromBot, err error) {
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
		if m, err = whc.NewEditMessage("", botsfw.MessageFormatText); err != nil {
			return
		}
		m.Keyboard = keyboardMarkup
	}

	return m, err
}

func getUserReportUrl(t strongo.SingleLocaleTranslator, submit string) string {
	switch t.Locale().Code5 {
	case strongo.LocalCodeRuRu:
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

func btnSubmitIdea(whc botsfw.WebhookContext, url string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.InlineKeyboardButton{
		Text: emoji.BULB_ICON + " " + whc.Translate(trans.COMMAND_TEXT_SUBMIT_AN_IDEA),
		URL:  url,
	}
}

func btnSubmitBug(whc botsfw.WebhookContext, url string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.InlineKeyboardButton{
		Text: emoji.ERROR_ICON + " " + whc.Translate(trans.COMMAND_TEXT_REPORT_A_BUG),
		URL:  url,
	}
}

const ADS_COMMAND = "ads"

var AdsCommand = botsfw.Command{
	Code:     ADS_COMMAND,
	Icon:     emoji.NEWSPAPER_ICON,
	Commands: []string{emoji.NEWSPAPER_ICON, "/ads", "/реклама"},
	Title:    trans.COMMAND_TEXT_HELP,
	Titles:   map[string]string{bots.ShortTitle: ""},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
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
