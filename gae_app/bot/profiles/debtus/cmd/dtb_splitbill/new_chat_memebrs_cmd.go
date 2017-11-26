package dtb_splitbill

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/log"
)

const NEW_CHAT_MEMBERS_COMMAND = "new-chat-members"

var NewChatMembersCommand = bots.Command{
	Code: NEW_CHAT_MEMBERS_COMMAND,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		log.Debugf(whc.Context(), "NewChatMembersCommand.Action()")
		if whc.IsInGroup() {
			log.Warningf(whc.Context(), "Leaving chat as @DebtusBot does not support group chats")
			m.BotMessage = telegram_bot.LeaveChat{}
		}
		return
	},
}
