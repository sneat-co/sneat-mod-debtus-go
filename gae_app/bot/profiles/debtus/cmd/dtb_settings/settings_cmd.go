package dtb_settings

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"net/url"
)

var SettingsCommand = shared_all.SettingsCommandTemplate

func init() {
	SettingsCommand.Action = func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		return shared_all.SettingsMainAction(whc)
	}
	SettingsCommand.CallbackAction = func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		return shared_all.SettingsMainAction(whc)
	}
}
