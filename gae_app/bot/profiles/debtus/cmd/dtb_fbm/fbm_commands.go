package dtb_fbm

import (
	"github.com/strongo/log"
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"net/url"
)

var FbmGetStartedCommand = bots.Command{ // TODO: Move command to other package?
	Code: "fbm-get-started",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "FbmGetStartedCommand.CallbackAction() => callbackUrl: %v", callbackUrl)
		//m.Text = "Welcome!"
		m.FbmAttachment = &fbm_api.RequestAttachment{
			Type: fbm_api.RequestAttachmentTypeTemplate,
		}

		if whc.ChatEntity().GetPreferredLanguage() == "" {
			m.FbmAttachment.Payload = askLanguageCard(whc)
		} else {
			m.FbmAttachment.Payload = fbm_api.NewGenericTemplate(
				welcomeCard(whc),
				debtsCard(whc),
				billsCard(whc),
				aboutCard(whc),
				linkAccountsCard(whc),
			)
		}
		return
	},
}

var FbmMainMenuCommand = bots.Command{
	Code: "fbm-main-menu",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "FbmMainMenuCommand.CallbackAction() => callbackUrl: %v", callbackUrl)

		m.FbmAttachment = &fbm_api.RequestAttachment{
			Type: fbm_api.RequestAttachmentTypeTemplate,
			Payload: fbm_api.NewGenericTemplate(
				mainMenuCard(whc),
				debtsCard(whc),
				billsCard(whc),
				aboutCard(whc),
				//linkAccountsCard(whc),
			),
		}
		return
	},
}

var FbmDebtsCommand = bots.Command{
	Code: "fbm-debts",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "FbmDebtsCommand.CallbackAction() => callbackUrl: %v", callbackUrl)

		m.FbmAttachment = &fbm_api.RequestAttachment{
			Type: fbm_api.RequestAttachmentTypeTemplate,
			Payload: fbm_api.NewGenericTemplate(
				debtsCard(whc),
			),
		}

		return
	},
}

var FbmBillsCommand = bots.Command{
	Code: "fbm-bills",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "FbmBillsCommand.CallbackAction() => callbackUrl: %v", callbackUrl)
		//m.Text = "Welcome!"
		m.FbmAttachment = &fbm_api.RequestAttachment{
			Type: fbm_api.RequestAttachmentTypeTemplate,
			Payload: fbm_api.NewGenericTemplate(
				billsCard(whc),
			),
		}

		return
	},
}

var FbmSettingsCommand = bots.Command{
	Code: "fbm-settings",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "FbmSettingsCommand.CallbackAction() => callbackUrl: %v", callbackUrl)
		m.FbmAttachment = &fbm_api.RequestAttachment{
			Type: fbm_api.RequestAttachmentTypeTemplate,
			Payload: fbm_api.NewGenericTemplate(
				settingsCard(whc),
			),
		}
		return
	},
}
