package dtb_fbm

import (
	"net/url"

	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

var FbmGetStartedCommand = bots.Command{ // TODO: Move command to other package?
	Code: "fbm-get-started",
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		log.Debugf(c, "FbmGetStartedCommand.CallbackAction() => callbackUrl: %v", callbackUrl)
		//m.Text = "Welcome!"
		m.FbmAttachment = &fbmbotapi.RequestAttachment{
			Type: fbmbotapi.RequestAttachmentTypeTemplate,
		}

		if whc.ChatEntity().GetPreferredLanguage() == "" {
			m.FbmAttachment.Payload = askLanguageCard(whc)
		} else {
			m.FbmAttachment.Payload = fbmbotapi.NewGenericTemplate(
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

		m.FbmAttachment = &fbmbotapi.RequestAttachment{
			Type: fbmbotapi.RequestAttachmentTypeTemplate,
			Payload: fbmbotapi.NewGenericTemplate(
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

		m.FbmAttachment = &fbmbotapi.RequestAttachment{
			Type: fbmbotapi.RequestAttachmentTypeTemplate,
			Payload: fbmbotapi.NewGenericTemplate(
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
		m.FbmAttachment = &fbmbotapi.RequestAttachment{
			Type: fbmbotapi.RequestAttachmentTypeTemplate,
			Payload: fbmbotapi.NewGenericTemplate(
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
		m.FbmAttachment = &fbmbotapi.RequestAttachment{
			Type: fbmbotapi.RequestAttachmentTypeTemplate,
			Payload: fbmbotapi.NewGenericTemplate(
				settingsCard(whc),
			),
		}
		return
	},
}
