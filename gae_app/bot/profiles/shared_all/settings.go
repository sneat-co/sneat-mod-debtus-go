package shared_all

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
)

const SettingsCommandCode = "settings"

func BackToSettingsAction(whc bots.WebhookContext, messageText string) (m bots.MessageFromBot, err error) {
	if messageText == "" {
		messageText = whc.Translate(trans.MESSAGE_TEXT_SETTINGS)
	} else {
		messageText += "\n\n" + whc.Translate(trans.MESSAGE_TEXT_SETTINGS)
	}
	m = whc.NewMessage(messageText)
	m.IsEdit = whc.InputType() == bots.WebhookInputCallbackQuery
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.CommandText(trans.COMMAND_TEXT_LANGUAGE, emoji.EARTH_ICON),
				CallbackData: SettingsLocaleListCallbackPath,
			},
		},
	)
	return m, err
}
