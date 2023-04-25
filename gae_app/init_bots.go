package gaeapp

import (
	"context"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/julienschmidt/httprouter"
	"github.com/sneat-co/debtstracker-go/gae_app/bot"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/platforms/tgbots"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/collectus"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/debtus"
	"github.com/sneat-co/debtstracker-go/gae_app/bot/profiles/splitus"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
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
