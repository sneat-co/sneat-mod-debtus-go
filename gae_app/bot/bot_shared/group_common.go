package bot_shared

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
)

func GroupCallbackCommandData(command string, groupID string) string {
	return command + "?group=" + groupID
}

type GroupAction func(whc bots.WebhookContext, group models.Group) (m bots.MessageFromBot, err error)
type GroupCallbackAction func(whc bots.WebhookContext, callbackUrl *url.URL, group models.Group) (m bots.MessageFromBot, err error)

func GroupCallbackCommand(code string, f GroupCallbackAction) bots.Command {
	return bots.NewCallbackCommand(code,
		func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
			var group models.Group
			if group, err = GetGroup(whc, callbackUrl); err != nil {
				return
			}
			return f(whc, callbackUrl, group)
		},
	)
}

func NewGroupAction(f GroupAction) bots.CommandAction {
	return func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		var group models.Group
		if group, err = GetGroup(whc, nil); err != nil {
			return
		}
		return f(whc, group)
	}
}

func NewGroupCallbackAction(f GroupCallbackAction) bots.CallbackAction {
	return func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		var group models.Group
		if group, err = GetGroup(whc, nil); err != nil {
			return
		}
		return f(whc, callbackUrl, group)
	}
}
