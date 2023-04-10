package splitus

import (
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
)

const LEAVE_GROUP_COMMAND = "leave-group"

var leaveGroupCommand = shared_group.GroupCallbackCommand(LEAVE_GROUP_COMMAND,
	func(whc botsfw.WebhookContext, _ *url.URL, group models.Group) (m botsfw.MessageFromBot, err error) {
		return
	},
)
