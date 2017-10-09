package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/decimal"
	"golang.org/x/net/context"
	"net/url"
)

const GROUP_MEMBERS_COMMAND = "group-members"

var groupMembersCommand = bots.Command{
	Code:     GROUP_MEMBERS_COMMAND,
	Commands: []string{"/members"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return showGroupMembers(whc, models.Group{}, false)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackURL *url.URL) (m bots.MessageFromBot, err error) {
		var group models.Group
		if group, err = bot_shared.GetGroup(whc); err != nil {
			err = nil
			return
		}
		return showGroupMembers(whc, group, true)
	},
}

func groupMembersCard(
	c context.Context,
	t strongo.SingleLocaleTranslator,
	group models.Group,
	selectedMemberID int64,
) (text string, err error) {
	var buffer bytes.Buffer
	buffer.WriteString(t.Translate(trans.MESSAGE_TEXT_MEMBERS_CARD_TITLE, group.MembersCount) + "\n\n")

	if group.GroupEntity == nil {
		if group, err = dal.Group.GetGroupByID(c, group.ID); err != nil {
			return
		}
	}

	if group.MembersCount > 0 {
		members := group.GetGroupMembers()
		if len(members) == 0 {
			msg := fmt.Sprintf("ERROR: group.MembersCount:%d != 0 && len(members) == 0", group.MembersCount)
			buffer.WriteString("\n" + msg + "\n")
			log.Errorf(c, msg)
		}

		splitMode := group.GetSplitMode()

		var totalShares int

		if splitMode != models.SplitModeEqually {
			totalShares = group.TotalShares()
		}

		for i, member := range members {
			if member.TgUserID == 0 {
				fmt.Fprintf(&buffer, `  %d. %v`, i+1, member.Name) // TODO: Do a proper padding with 0 on left of #
			} else {
				fmt.Fprintf(&buffer, `  %d. <a href="tg://user?id=%d">%v</a>`, i+1, member.TgUserID, member.Name)
			}
			if splitMode != models.SplitModeEqually {
				fmt.Fprintf(&buffer, " (%d%%)", decimal.Decimal64p2(member.Shares*100/totalShares))
			}
			fmt.Fprintln(&buffer)
		}
	}

	buffer.WriteString("\n" + t.Translate(trans.MESSAGE_TEXT_MEMBERS_CARD_FOOTER))

	return buffer.String(), nil
}

func showGroupMembers(whc bots.WebhookContext, group models.Group, isEdit bool) (m bots.MessageFromBot, err error) {

	if group.GroupEntity == nil {
		if group, err = bot_shared.GetGroup(whc); err != nil {
			return
		}
	}

	c := whc.Context()

	if m.Text, err = groupMembersCard(c, whc, group, 0); err != nil {
		return
	}

	m.Format = bots.MessageFormatHTML
	tgKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BUTTON_TEXT_JOIN),
				CallbackData: bot_shared.JOIN_GROUP_COMMAND,
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.Translate(trans.BUTTON_TEXT_SPLIT_MODE, whc.Translate(string(group.GetSplitMode()))),
				CallbackData: bot_shared.GroupCallbackCommandData(GROUP_SPLIT_COMMAND, group.ID),
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(
				emoji.CONTACTS_ICON+" "+whc.Translate(trans.COMMAND_TEXT_INVITE_MEMBER),
				bot_shared.GroupCallbackCommandData(bot_shared.JOIN_GROUP_COMMAND, group.ID),
			),
		},
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonSwitchInlineQueryCurrentChat(
				emoji.CLIPBOARD_ICON+whc.Translate(trans.COMMAND_TEXT_NEW_BILL),
				"",
			),
		},
	)
	m.Keyboard = tgKeyboard
	m.IsEdit = isEdit
	return
}
