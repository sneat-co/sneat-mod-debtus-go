package dtb_invite

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_general"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

const INVITE_COMMAND = "invite"

var InviteCommand = bots.Command{
	Code:     INVITE_COMMAND,
	Commands: []string{dtb_general.INVITES_SHOT_COMMAND, "/Пригласить_друга", "/invite"},
	Replies: []bots.Command{
		AskInviteAddressTelegramCommand,
		AskInviteAddressEmailCommand,
		AskInviteAddressSmsCommand,
	},
	Action: func(whc bots.WebhookContext) (bots.MessageFromBot, error) {
		m := whc.NewMessageByCode(trans.MESSAGE_TEXT_ABOUT_INVITES)
		m.Keyboard = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(AskInviteAddressTelegramCommand.DefaultTitle(whc), "/invite"),
				},
				{
					{
						Text:         AskInviteAddressSmsCommand.DefaultTitle(whc),
						CallbackData: "invite?by=sms",
					},
					{
						Text:         AskInviteAddressEmailCommand.DefaultTitle(whc),
						CallbackData: "invite?by=email",
					},
				},
			},
		}
		whc.ChatEntity().SetAwaitingReplyTo(INVITE_COMMAND)
		return m, nil
	},
}
