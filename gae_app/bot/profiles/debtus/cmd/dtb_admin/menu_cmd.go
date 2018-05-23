package dtb_admin

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_invite"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

var AdminCommand = bots.Command{
	Code:     "admin",
	Commands: []string{"/admin"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		m = whc.NewMessage("Admin menu")
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{Text: "Create mass invite", CallbackData: dtb_invite.CREATE_MASS_INVITE_CALLBACK_PATH},
			},
		)
		return
	},
}
