package splitus

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"fmt"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"golang.org/x/net/context"
	"github.com/DebtsTracker/translations/trans"
	"bitbucket.com/debtstracker/gae_app/bot/bot_shared"
)

const GROUP_SPLIT_COMMAND = "group-split"

var groupSplitCommand = bot_shared.GroupCallbackCommand(GROUP_SPLIT_COMMAND,
	func(whc bots.WebhookContext, callbackURL *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		return editSplitCallbackAction(
			whc, callbackURL,
			bot_shared.GroupCallbackCommandData(GROUP_SPLIT_COMMAND, group.ID),
			bot_shared.GroupCallbackCommandData(GROUP_MEMBERS_COMMAND, group.ID),
			trans.MESSAGE_TEXT_ASK_HOW_TO_SPLIT_IN_GROP,
			group.GetMembers(),
			models.Amount{},
			nil,
			func(memberID string, addValue int) (member models.MemberJson, err error) {
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
							member = m.MemberJson
							return err
						}
					}
					return fmt.Errorf("member not found by ID: %d", member.ID)
				}, dal.CrossGroupTransaction)
				return
			},
		)
	},
)
