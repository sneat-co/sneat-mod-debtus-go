package splitus

import (
	"github.com/sneat-co/debtstracker-translations/emoji"
)

const menuCommandCode = "menu"

var menuCommand = botsfw.Command{
	Code:     menuCommandCode,
	Commands: []string{"/" + menuCommandCode},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		m.Text = whc.Translate(trans.SPLITUS_TG_COMMANDS)
		m.Format = botsfw.MessageFormatHTML
		setMainMenu(whc, &m)
		return
	},
}

func setMainMenu(whc botsfw.WebhookContext, m *botsfw.MessageFromBot) {
	m.Keyboard = tgbotapi.NewReplyKeyboard(
		[]tgbotapi.KeyboardButton{
			{Text: groupsCommand.TitleByKey(botsfw.DefaultTitle, whc)},
			{Text: billsCommand.TitleByKey(botsfw.DefaultTitle, whc)},
		},
		[]tgbotapi.KeyboardButton{
			{Text: emoji.SETTINGS_ICON + " " + whc.Translate(trans.COMMAND_TEXT_SETTING)},
			{Text: emoji.HELP_ICON + " " + whc.Translate(trans.COMMAND_TEXT_HELP)},
		},
	)
}
