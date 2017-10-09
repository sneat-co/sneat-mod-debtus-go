package bot_shared

import "github.com/strongo/bots-framework/core"

const CHAT_LEFT_COMMAND = "left-chat"

var leftChatCommand = bots.Command{
	Code: CHAT_LEFT_COMMAND,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return
	},
}
