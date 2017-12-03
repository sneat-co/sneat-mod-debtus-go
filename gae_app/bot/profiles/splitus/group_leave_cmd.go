package splitus

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_group"
)

const LEAVE_GROUP_COMMAND = "leave-group"

var leaveGroupCommand = shared_group.GroupCallbackCommand(LEAVE_GROUP_COMMAND,
	func(whc bots.WebhookContext, _ *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		return
	},
)
