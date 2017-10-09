package dtb_invite

import (
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

const (
	CREATE_MASS_INVITE_CALLBACK_PATH = "create-mass-invite"
	CREATE_MASS_INVITE_COMMAND_CODE  = CREATE_MASS_INVITE_CALLBACK_PATH
)

var CreateMassInviteCommand = bots.Command{
	Code:     CREATE_MASS_INVITE_COMMAND_CODE,
	Commands: []string{"/massinvite"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		m = whc.NewMessage("Admin menu")

		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{Text: "Create mass invite", URL: CREATE_MASS_INVITE_CALLBACK_PATH},
			},
		)
		return
	},
}
