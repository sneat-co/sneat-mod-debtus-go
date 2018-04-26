package shared_all

import (
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

const SettingsCommandCode = "settings"

var SettingsCommandTemplate = bots.Command{
	Code:     SettingsCommandCode,
	Commands: trans.Commands(trans.COMMAND_TEXT_SETTING, trans.COMMAND_SETTINGS, emoji.SETTINGS_ICON),
	Icon:     emoji.SETTINGS_ICON,
}

func SettingsMainAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	switch whc.BotPlatform().ID() {
	case telegram.PlatformID:
		m, _, err = SettingsMainTelegram(whc)
	default:
		err = errors.New("Unsupported platform")
	}
	return
}

func SettingsMainTelegram(whc bots.WebhookContext) (m bots.MessageFromBot, keyboard *tgbotapi.InlineKeyboardMarkup, err error) {
	m = whc.NewMessage(whc.Translate(trans.MESSAGE_TEXT_SETTINGS))
	m.IsEdit = whc.InputType() == bots.WebhookInputCallbackQuery
	keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.CommandText(trans.COMMAND_TEXT_LANGUAGE, emoji.EARTH_ICON),
				CallbackData: SettingsLocaleListCallbackPath,
			},
		},
	)
	m.Keyboard = keyboard
	return
}
