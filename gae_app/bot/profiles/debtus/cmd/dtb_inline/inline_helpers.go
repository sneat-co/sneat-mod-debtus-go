package dtb_inline

import (
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
)

func GetChooseLangInlineKeyboard(format string, currentLocaleCode5 string) []tgbotapi.InlineKeyboardButton {
	buttons := []tgbotapi.InlineKeyboardButton{}

	for code5, locale := range trans.SupportedLocalesByCode5 {
		if code5 != currentLocaleCode5 {
			buttons = append(buttons, tgbotapi.InlineKeyboardButton{
				Text:         locale.TitleWithIcon(),
				CallbackData: fmt.Sprintf(format, locale.Code5),
			})
		}
	}
	return buttons
}
