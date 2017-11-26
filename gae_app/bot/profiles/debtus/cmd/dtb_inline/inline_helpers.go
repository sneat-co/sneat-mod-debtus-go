package dtb_inline

import (
	"fmt"

	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
)

func GetChooseLangInlineKeyboard(format string, currentLocaleCode5 string) (kbRows [][]tgbotapi.InlineKeyboardButton) {
	kbRows = make([][]tgbotapi.InlineKeyboardButton, 0, len(trans.SupportedLocalesByCode5))

	for code5, locale := range trans.SupportedLocalesByCode5 {
		if code5 != currentLocaleCode5 {
			btnRow := []tgbotapi.InlineKeyboardButton{
				{
					Text:         locale.TitleWithIcon(),
					CallbackData: fmt.Sprintf(format, locale.Code5),
				},
			}
			kbRows = append(kbRows, btnRow)
		}
	}

	return
}
