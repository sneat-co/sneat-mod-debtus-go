package splitus

import (
	"fmt"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

const NEW_CHAT_MEMBERS_COMMAND = "new-chat-members"

var newChatMembersCommand = bots.Command{
	Code: NEW_CHAT_MEMBERS_COMMAND,
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		c := whc.Context()

		newMembersMessage := whc.Input().(bots.WebhookNewChatMembersMessage)

		newMembers := newMembersMessage.NewChatMembers()

		{ // filter out bots
			j := 0
			for _, member := range newMembers {
				if !member.IsBotUser() {
					newMembers[j] = member
					j += 1
				}
			}
			newMembers = newMembers[:j]
		}

		if len(newMembers) == 0 {
			return
		}

		var newUsers []facade.NewUser

		{ // Get or create related user records
			for _, chatMember := range newMembers {
				tgChatMember := chatMember.(tgbotapi.ChatMember)
				var botUser bots.BotUser
				if botUser, err = whc.GetBotUserById(c, tgChatMember.ID); err != nil {
					return
				}
				if botUser == nil {
					if botUser, err = whc.CreateBotUser(c, whc.GetBotCode(), chatMember); err != nil {
						return
					}
				}
				newUsers = append(newUsers, facade.NewUser{
					Name:       tgChatMember.GetFullName(),
					BotUser:    botUser,
					ChatMember: chatMember,
				})
			}
		}

		var group models.Group
		if group, err = shared_group.GetGroup(whc, nil); err != nil {
			return
		}
		if group, newUsers, err = facade.Group.AddUsersToTheGroupAndOutstandingBills(whc.Context(), group.ID, newUsers); err != nil {
			return
		}

		if len(newUsers) == 0 {
			return
		}

		createWelcomeMsg := func(member bots.WebhookActor) bots.MessageFromBot {
			m := whc.NewMessageByCode(trans.MESSAGE_TEXT_USER_JOINED_GROUP, member.GetFirstName())
			m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{
					{
						Text: whc.CommandText(trans.COMMAND_TEXT_SETTING, emoji.SETTINGS_ICON),
						URL:  fmt.Sprintf("https:/t.me/%v?start=group-%v", whc.GetBotCode(), group.ID),
					},
				},
			)

			return m
		}
		m = createWelcomeMsg(newUsers[0].ChatMember)

		if len(newUsers) > 1 {
			responder := whc.Responder()
			c := whc.Context()
			for _, newUser := range newUsers {
				if _, err = responder.SendMessage(c, createWelcomeMsg(newUser.ChatMember), bots.BotApiSendMessageOverHTTPS); err != nil {
					return
				}
			}
		}
		return
	},
}
