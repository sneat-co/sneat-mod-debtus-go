package splitus

import (
	"fmt"
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/bot_shared"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
)

var groupSplitCommand = bot_shared.GroupCallbackCommand(bot_shared.GROUP_SPLIT_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		c := whc.Context()

		members := group.GetMembers()
		billMembers := make([]models.BillMemberJson, len(members))
		for i, m := range members {
			billMembers[i].MemberJson = m
		}
		return editSplitCallbackAction(
			whc, callbackUrl,
			bot_shared.GroupCallbackCommandData(bot_shared.GROUP_SPLIT_COMMAND, group.ID),
			bot_shared.GroupCallbackCommandData(bot_shared.SETTINGS_COMMAND, group.ID),
			trans.MESSAGE_TEXT_ASK_HOW_TO_SPLIT_IN_GROP,
			billMembers,
			models.Amount{},
			nil,
			func(memberID string, addValue int) (member models.BillMemberJson, err error) {
				err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
					if group, err = dal.Group.GetGroupByID(c, group.ID); err != nil {
						return
					}
					members := group.GetGroupMembers()
					for i, m := range members {
						if m.ID == memberID {
							m.Shares += addValue
							if m.Shares < 0 {
								m.Shares = 0
							}
							members[i] = m
							group.SetGroupMembers(members)
							if err = dal.Group.SaveGroup(c, group); err != nil {
								return
							}
							member = models.BillMemberJson{MemberJson: m.MemberJson}
							return err
						}
					}
					return fmt.Errorf("member not found by ID: %v", member.ID)
				}, dal.CrossGroupTransaction)
				return
			},
		)
	},
)
