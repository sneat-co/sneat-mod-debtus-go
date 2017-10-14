package bot_shared

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
	"net/url"
)

func GroupCallbackCommandData(command string, groupID string) string {
	return command + "?group=" + groupID
}

func GroupCallbackCommand(code string, f func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error)) bots.Command {
	return bots.NewCallbackCommand(code,
		func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
			var group models.Group
			if group, err = GetGroup(whc); err != nil {
				return
			}
			return f(whc, callbackUrl, group)
		},
	)
}
