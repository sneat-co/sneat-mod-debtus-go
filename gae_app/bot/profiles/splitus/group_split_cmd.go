package splitus

import (
	"fmt"
	"github.com/crediterra/money"
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

const groupSplitCommandCode = "group-split"

var groupSplitCommand = shared_group.GroupCallbackCommand(groupSplitCommandCode,
	func(whc botsfw.WebhookContext, callbackUrl *url.URL, group models.Group) (m botsfw.MessageFromBot, err error) {
		c := whc.Context()

		members := group.GetMembers()
		billMembers := make([]models.BillMemberJson, len(members))
		for i, m := range members {
			billMembers[i].MemberJson = m
		}
		return editSplitCallbackAction(
			whc, callbackUrl,
			"",
			shared_group.GroupCallbackCommandData(groupSplitCommandCode, group.ID),
			shared_group.GroupCallbackCommandData(shared_all.SettingsCommandCode, group.ID),
			trans.MESSAGE_TEXT_ASK_HOW_TO_SPLIT_IN_GROP,
			billMembers,
			money.Amount{},
			nil,
			func(memberID string, addValue int) (member models.BillMemberJson, err error) {
				err = dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
					if group, err = dtdal.Group.GetGroupByID(c, group.ID); err != nil {
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
							if err = dtdal.Group.SaveGroup(c, group); err != nil {
								return
							}
							member = models.BillMemberJson{MemberJson: m.MemberJson}
							return err
						}
					}
					return fmt.Errorf("member not found by ID: %v", member.ID)
				}, dtdal.CrossGroupTransaction)
				return
			},
		)
	},
)
