package shared_all

const CHAT_LEFT_COMMAND = "left-chat"

var leftChatCommand = botsfw.Command{
	Code: CHAT_LEFT_COMMAND,
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		return
	},
}
