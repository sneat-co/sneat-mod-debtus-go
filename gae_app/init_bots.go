package gaeapp

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/platforms/tgbots"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/collectus"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/debtus"
	"bitbucket.org/asterus/debtstracker-server/gae_app/bot/profiles/splitus"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"context"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/julienschmidt/httprouter"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/app"
)

func newTranslator(c context.Context) strongo.Translator {
	return strongo.NewMapTranslator(c, trans.TRANS)
}

func InitBots(httpRouter *httprouter.Router, botHost botsfw.BotHost, appContext botsfw.BotAppContext) {

	driver := botsfw.NewBotDriver( // Orchestrate requests to appropriate handlers
		botsfw.AnalyticsSettings{GaTrackingID: common.GA_TRACKING_ID}, // TODO: Refactor to list of analytics providers
		appContext, // Holds User entity kind name, translator, etc.
		botHost,    // Defines how to create context.Context, HttpClient, DB, etc...
		"Please report any issues to @DebtsTrackerGroup", // Is it wrong place? Router has similar.
	)

	driver.RegisterWebhookHandlers(httpRouter, "/bot",
		telegram.NewTelegramWebhookHandler(
			telegramBotsWithRouter, // Maps of bots by code, language, token, etc...
			newTranslator,          // Creates translator that gets a context.Context (for logging purpose)
		),
		//viber.NewViberWebhookHandler(
		//	viberbots.Bots,
		//	newTranslator,
		//),
		//fbm.NewFbmWebhookHandler(
		//	fbmbots.Bots,
		//	newTranslator,
		//),
	)
}

func telegramBotsWithRouter(c context.Context) botsfw.SettingsBy {
	return tgbots.Bots(dtdal.HttpAppHost.GetEnvironment(c, nil), func(profile string) botsfw.WebhooksRouter {
		switch profile {
		case bot.ProfileDebtus:
			return debtus.Router
		case bot.ProfileSplitus:
			return splitus.Router
		case bot.ProfileCollectus:
			return collectus.Router
		default:
			panic("Unknown bot profile: " + profile)
		}
	})
}
