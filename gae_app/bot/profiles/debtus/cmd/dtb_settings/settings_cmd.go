package dtb_settings

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/shared_all"
	"github.com/strongo/bots-framework/core"
	"net/url"
)

var SettingsCommand = shared_all.SettingsCommandTemplate

func init() {
	SettingsCommand.Action = func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return shared_all.SettingsMainAction(whc)
	}
	SettingsCommand.CallbackAction = func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return shared_all.SettingsMainAction(whc)
	}
}
