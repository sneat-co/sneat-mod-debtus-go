package tgbots

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-fw-store/botsfwmodels"
	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/common"
	"github.com/strongo/app"
	"github.com/strongo/i18n"
)

var _bots botsfw.SettingsBy

const DefaultLocale = i18n.LocaleCodeEnUS

const DebtusBotToken = "467112035:AAG9Hij0ofnI6GGXyuc6zol0F4XGQ4OK5Tk"

func newTelegramBot(
	mode strongo.Environment,
	botProfile botsfw.BotProfile,
	code, gaToken string,
	locale i18n.Locale,
) botsfw.BotSettings {
	return telegram.NewTelegramBot(mode, botProfile, code, "", "", "", gaToken, i18n.LocaleEnUS, nil, nil)
}

func Bots(environment strongo.Environment, router func(profile string) botsfw.WebhooksRouter) botsfw.SettingsBy { //TODO: Consider to do pre-deployment replace
	newBotChatData := func() botsfwmodels.BotChatData {
		return nil
	}

	newBotUserData := func() botsfwmodels.BotUserData {
		return nil
	}
	newAppUserData := func() botsfwmodels.AppUserData {
		return nil
	}
	getAppUserByID := func(c context.Context, tx dal.ReadSession, botID, appUserID string) (appUser record.DataWithID[string, botsfwmodels.AppUserData], err error) {
		//var userID int64
		//userID, err = strconv.ParseInt(appUserID, 10, 64)
		//if err != nil {
		//	return appUser, fmt.Errorf("failed to parse appUserID as int64: %w", err)
		//}
		appUserData := newAppUserData()
		d := record.NewDataWithID(appUserID, dal.NewKeyWithID("Users", appUserID), appUserData)
		appUser = d

		return appUser, nil
	}

	debtusBotProfile := botsfw.NewBotProfile("debtus", nil, newBotChatData, newBotUserData, newAppUserData, getAppUserByID, i18n.LocaleEnUS, nil)
	splitusBotProfile := botsfw.NewBotProfile("splitus", nil, newBotChatData, newBotUserData, newAppUserData, getAppUserByID, i18n.LocaleEnUS, nil)
	collectusBotProfile := botsfw.NewBotProfile("collectus", nil, newBotChatData, newBotUserData, newAppUserData, getAppUserByID, i18n.LocaleEnUS, nil)

	if len(_bots.ByCode) == 0 {
		//log.Debugf(c, "Bots() => hostname:%v, environment:%v:%v", hostname, environment, strongo.EnvironmentNames[environment])
		switch environment {
		case strongo.EnvProduction:
			_bots = botsfw.NewBotSettingsBy( // Production bots
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerBot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				newTelegramBot(strongo.EnvProduction, splitusBotProfile, "SplitusBot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				newTelegramBot(strongo.EnvProduction, collectusBotProfile, "CollectusBot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerRuBot", common.GA_TRACKING_ID, i18n.LocaleRuRu),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerFaBot", common.GA_TRACKING_ID, i18n.LocalesByCode5[i18n.LocaleCodeFaIR]),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerItBot", common.GA_TRACKING_ID, i18n.LocaleItIt),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerFrBot", common.GA_TRACKING_ID, i18n.LocaleFrFr),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerDeBot", common.GA_TRACKING_ID, i18n.LocaleDeDe),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerPLbot", common.GA_TRACKING_ID, i18n.LocalePlPl),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerPtBot", common.GA_TRACKING_ID, i18n.LocalePtBr),
				newTelegramBot(strongo.EnvProduction, debtusBotProfile, "DebtsTrackerEsBot", common.GA_TRACKING_ID, i18n.LocalePtBr),
			)
		case strongo.EnvDevTest:
			_bots = botsfw.NewBotSettingsBy( // Development bots
				newTelegramBot(strongo.EnvDevTest, debtusBotProfile, "DebtsTrackerDev1Bot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				newTelegramBot(strongo.EnvDevTest, debtusBotProfile, "DebtsTrackerDev1RuBot", common.GA_TRACKING_ID, i18n.LocaleRuRu),
				//telegram.NewTelegramBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev2RuBot", "360514041:AAFXuT0STHBD9cOn1SFmKzTYDmalP0Rz-7M", "", "", common.GA_TRACKING_ID, i18n.LocalesByCode5[i18n.LocalCodeRuRu]),
			)
		case strongo.EnvStaging:
			_bots = botsfw.NewBotSettingsBy( // Staging bots
				newTelegramBot(strongo.EnvStaging, debtusBotProfile, "DebtsTrackerSt1Bot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
			)
		case strongo.EnvLocal:
			_bots = botsfw.NewBotSettingsBy( // Staging bots
				newTelegramBot(strongo.EnvLocal, debtusBotProfile, "DebtsTrackerLocalBot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				newTelegramBot(strongo.EnvLocal, splitusBotProfile, "SplitusLocalBot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				newTelegramBot(strongo.EnvLocal, collectusBotProfile, "CollectusLocalBot", common.GA_TRACKING_ID, i18n.LocaleEnUS),
			)
		case strongo.EnvUnknown:
			// Pass for unit tests?
		default:
			panic(fmt.Sprintf("Unknown environment => %v:%v", environment, strongo.EnvironmentNames[environment]))
		}
	}
	return _bots
}

func GetBotSettingsByLang(environment strongo.Environment, profile, lang string) (botsfw.BotSettings, error) {
	botSettingsBy := Bots(environment, nil)
	for _, bs := range botSettingsBy.ByCode {
		if bs.Profile.ID() == profile && bs.Locale.Code5 == lang {
			return *bs, nil
		}
	}
	for _, bs := range botSettingsBy.ByCode {
		if bs.Profile.ID() == profile && bs.Locale.Code5 == DefaultLocale {
			return *bs, nil
		}
	}
	return botsfw.BotSettings{}, fmt.Errorf("no bot setting for both %s & %s locales", lang, DefaultLocale)
}
