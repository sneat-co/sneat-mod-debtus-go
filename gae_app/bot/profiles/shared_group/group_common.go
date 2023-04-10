package shared_group

import (
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
)

func GroupCallbackCommandData(command string, groupID string) string {
	return command + "?group=" + groupID
}

type GroupAction func(whc botsfw.WebhookContext, group models.Group) (m botsfw.MessageFromBot, err error)
type GroupCallbackAction func(whc botsfw.WebhookContext, callbackUrl *url.URL, group models.Group) (m botsfw.MessageFromBot, err error)

func GroupCallbackCommand(code string, f GroupCallbackAction) botsfw.Command {
	return botsfw.NewCallbackCommand(code,
		func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
			var group models.Group
			if group, err = GetGroup(whc, callbackUrl); err != nil {
				return
			}
			return f(whc, callbackUrl, group)
		},
	)
}

func NewGroupAction(f GroupAction) botsfw.CommandAction {
	return func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		var group models.Group
		if group, err = GetGroup(whc, nil); err != nil {
			return
		}
		return f(whc, group)
	}
}

func NewGroupCallbackAction(f GroupCallbackAction) botsfw.CallbackAction {
	return func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		var group models.Group
		if group, err = GetGroup(whc, nil); err != nil {
			return
		}
		return f(whc, callbackUrl, group)
	}
}
