package dtb_invite

const (
	CREATE_MASS_INVITE_CALLBACK_PATH = "create-mass-invite"
	CREATE_MASS_INVITE_COMMAND_CODE  = CREATE_MASS_INVITE_CALLBACK_PATH
)

var CreateMassInviteCommand = botsfw.Command{
	Code:     CREATE_MASS_INVITE_COMMAND_CODE,
	Commands: []string{"/massinvite"},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		m = whc.NewMessage("Admin menu")

		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				{Text: "Create mass invite", URL: CREATE_MASS_INVITE_CALLBACK_PATH},
			},
		)
		return
	},
}
