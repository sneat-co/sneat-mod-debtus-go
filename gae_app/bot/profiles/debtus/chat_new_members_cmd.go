package debtus

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
)

const NEW_CHAT_MEMBERS_COMMAND = "new-chat-members"

var newChatMembersCommand = bots.Command{
	Code: NEW_CHAT_MEMBERS_COMMAND,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		if whc.IsInGroup() {
			log.Warningf(whc.Context(), "Leaving chat as @DebtsTrackerBot does not support group chats")
			m.BotMessage = telegram_bot.LeaveChat{}
		}
		return
	},
}
