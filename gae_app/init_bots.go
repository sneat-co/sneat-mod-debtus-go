package gae_app

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/fbm"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/viber"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/collectus"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/splitus"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/app/gaestandard"
	"github.com/DebtsTracker/translations/trans"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/fbm"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-framework/platforms/viber"
	"context"
)

func newTranslator(c context.Context) strongo.Translator {
	return strongo.NewMapTranslator(c, trans.TRANS)
}

func InitBots(httpRouter *httprouter.Router, botHost bots.BotHost, appContext bots.BotAppContext) {

	driver := bots.NewBotDriver( // Orchestrate requests to appropriate handlers
		bots.AnalyticsSettings{GaTrackingID: common.GA_TRACKING_ID}, // TODO: Refactor to list of analytics providers
		appContext, // Holds User entity kind name, translator, etc.
		botHost,    // Defines how to create context.Context, HttpClient, DB, etc...
		"Please report any issues to @DebtsTrackerGroup", // Is it wrong place? Router has similar.
	)

	driver.RegisterWebhookHandlers(httpRouter, "/bot",
		telegram.NewTelegramWebhookHandler(
			telegramBotsWithRouter, // Maps of bots by code, language, token, etc...
			newTranslator,          // Creates translator that gets a context.Context (for logging purpose)
		),
		viber.NewViberWebhookHandler(
			viber.Bots,
			newTranslator,
		),
		fbm.NewFbmWebhookHandler(
			fbm.Bots,
			newTranslator,
		),
	)
}

func telegramBotsWithRouter(c context.Context) bots.SettingsBy {
	return telegram.Bots(gaestandard.GetEnvironment(c), func(profile string) bots.WebhooksRouter {
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
