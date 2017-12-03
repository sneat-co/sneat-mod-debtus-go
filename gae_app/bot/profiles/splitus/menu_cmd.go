package splitus

import (
	"github.com/strongo/bots-framework/core"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
)

const menuCommandCode = "menu"

var menuCommand = bots.Command{
	Code: menuCommandCode,
	Commands: []string{"/"+menuCommandCode},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		m.Text = whc.Translate(trans.SPLITUS_TG_COMMANDS)
		m.Format = bots.MessageFormatHTML
		m.Keyboard = telegramMainMenu(whc)
		return
	},
}

func telegramMainMenu(whc bots.WebhookContext) bots.Keyboard {
	return tgbotapi.NewReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_SETTING)},
		},
		[]tgbotapi.KeyboardButton{
			{Text: whc.Translate(trans.COMMAND_TEXT_HELP)},
		},
	)
}
