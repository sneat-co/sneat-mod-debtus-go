package splitus

import (
	"bytes"
	"fmt"
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/DebtsTracker/translations/emoji"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
)

const GroupMembersCommandCode = "group-members"

var groupMembersCommand = bots.Command{
	Code:     GroupMembersCommandCode,
	Commands: []string{"/members"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return showGroupMembers(whc, models.Group{}, false)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		var group models.Group
		if group, err = shared_group.GetGroup(whc, callbackUrl); err != nil {
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
		if group, err = dtdal.Group.GetGroupByID(c, group.ID); err != nil {
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
			if member.TgUserID == "" {
				fmt.Fprintf(&buffer, `  %d. %v`, i+1, member.Name) // TODO: Do a proper padding with 0 on left of #
			} else {
				fmt.Fprintf(&buffer, `  %d. <a href="tg://user?id=%v">%v</a>`, i+1, member.TgUserID, member.Name)
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
		if group, err = shared_group.GetGroup(whc, nil); err != nil {
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
				CallbackData: joinGroupCommanCode,
			},
		},
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonSwitchInlineQuery(
				emoji.CONTACTS_ICON+" "+whc.Translate(trans.COMMAND_TEXT_INVITE_MEMBER),
				shared_group.GroupCallbackCommandData(joinGroupCommanCode, group.ID),
			),
		},
		[]tgbotapi.InlineKeyboardButton{
			{
				Text:         whc.CommandText(trans.COMMAND_TEXT_SETTING, emoji.SETTINGS_ICON),
				CallbackData: shared_group.GroupCallbackCommandData(shared_all.SettingsCommandCode, group.ID),
			},
		},
	)
	m.Keyboard = tgKeyboard
	m.IsEdit = isEdit
	return
}
