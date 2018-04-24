package splitus

import (
	"bytes"
	"fmt"
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
)

const GROUP_BALANCE_COMMAND = "group-balance"

var groupBalanceCommand = bots.Command{
	Code:     GROUP_BALANCE_COMMAND,
	Commands: []string{"/balance"},
	Action:   shared_group.NewGroupAction(groupBalanceAction),
	CallbackAction: shared_group.NewGroupCallbackAction(func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		return groupBalanceAction(whc, group)
	}),
}

func groupBalanceAction(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error) {
	var buf bytes.Buffer
	writeMembers := func(members []models.GroupMemberJson) {
		for i, m := range members {
			fmt.Fprintf(&buf, " %d. %v:", i+1, m.Name)
			for currency, amount := range m.Balance {
				if amount < 0 {
					amount *= -1
				}
				fmt.Fprintf(&buf, " %v %v,", amount, currency)
			}
			buf.Truncate(buf.Len() - 1)
			buf.WriteString("\n")
		}
	}
	groupMembers := group.GetGroupMembers()
	sponsors, debtors := getGroupSponsorsAndDebtors(groupMembers)

	buf.WriteString(whc.Translate(trans.MT_GROUP_LABEL, group.Name))
	buf.WriteString("\n")

	buf.WriteString("\n")
	buf.WriteString(whc.Translate(trans.MT_SPONSORS_HEADER))
	buf.WriteString("\n")
	writeMembers(sponsors)

	buf.WriteString("\n")
	buf.WriteString(whc.Translate(trans.MT_DEBTORS_HEADER))
	buf.WriteString("\n")
	writeMembers(debtors)

	m.Text = buf.String()
	m.Format = bots.MessageFormatHTML
	m.IsEdit = whc.Input().InputType() == bots.WebhookInputCallbackQuery

	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			{
				Text: "Settle up",
				URL:  shared_all.StartBotLink(whc.GetBotCode(), SettleGroupAskForCounterpartyCommandCode, "group="+group.ID),
			},
		},
	)
	return
}

func getGroupSponsorsAndDebtors(members []models.GroupMemberJson, excludeMemberIDs ...string) (sponsors, debtors []models.GroupMemberJson) {
	sponsors = make([]models.GroupMemberJson, 0, len(members))
	debtors = make([]models.GroupMemberJson, 0, len(members))

	for _, m := range members {
		for _, id := range excludeMemberIDs {
			if m.ID == id {
				continue
			}
		}
		for _, v := range m.Balance {
			if v > 0 {
				sponsors = append(sponsors, m)
			} else if v < 0 {
				debtors = append(debtors, m)
			}
		}
	}
	return
}

//func removeGroupMemberByID(members []models.GroupMemberJson, excludeMemberID string) ([]models.GroupMemberJson) {
//	for i, m := range members {
//		if m.ID == excludeMemberID {
//			return append(members[:i], members[i+1:]...)
//		}
//	}
//	return models.GroupMemberJson{}, members
//}
