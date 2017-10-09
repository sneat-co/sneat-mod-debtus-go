package bot_shared

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
)

const LEAVE_GROUP_COMMAND = "leave-group"

var leaveGroupCommand = GroupCallbackCommand(LEAVE_GROUP_COMMAND,
	func(whc bots.WebhookContext, _ *url.URL, group models.Group) (m bots.MessageFromBot, err error) {
		return
	},
)


