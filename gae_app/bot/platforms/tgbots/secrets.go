package tgbots

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/i18n"
	"strings"

	"github.com/bots-go-framework/bots-fw-telegram"
	"github.com/sneat-co/debtstracker-go/gae_app/bot"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
	"github.com/strongo/app"
)

var _bots botsfw.SettingsBy

const DEFAULT_LOCALE = i18n.LocaleCodeEnUS

const DebtusBotToken = "467112035:AAG9Hij0ofnI6GGXyuc6zol0F4XGQ4OK5Tk"

func Bots(environment strongo.Environment, router func(profile string) botsfw.WebhooksRouter) botsfw.SettingsBy { //TODO: Consider to do pre-deployment replace
	if len(_bots.ByCode) == 0 || (!_bots.HasRouter && router != nil) {
		//log.Debugf(c, "Bots() => hostname:%v, environment:%v:%v", hostname, environment, strongo.EnvironmentNames[environment])
		switch environment {
		case strongo.EnvProduction:
			_bots = botsfw.NewBotSettingsBy(router,
				// Production bots
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileCollectus, "CollectusBot", "458860316:AAFk_hOXK5vFWu43jp4apWgQjmHHv87CU9E", "", "", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileSplitus, "SplitusBot", "345328965:AAHmM7rUCwiPBlVIv-IfhrWhYIUVSHerkpg", "", "", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerBot", "209808950:AAHEwdBVtVIhKZhieTCP6zdbVkTROoj0fyA", "284685063:TEST:Njc4MWQ2NzlmMDAx", "350862534:LIVE:ZjAzOWE3ODg5OWMy", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerRuBot", "218446201:AAGyvWHuodNYT8kgbR_701m6y8Xg5D9iTSA", "284685063:TEST:MDg3NzM5ZTUxMTNk	", "350862534:LIVE:MGM1ODY0N2Q2ZDM5", common.GA_TRACKING_ID, i18n.LocaleRuRu),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerFaBot", "182148042:AAFHD7MfWr5CLjGczaiqsx-Oo6msoR_5JfM", "", "", common.GA_TRACKING_ID, i18n.LocalesByCode5[i18n.LocaleCodeFaIR]),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerItBot", "143800015:AAFrLrjyKCIqVFE0YsdZghYtDVmiLpa_P_A", "84685063:TEST:Zjg1ZTIxYzEyNTQ3", "350862534:LIVE:ZmRhMWRhOWZiOWIx", common.GA_TRACKING_ID, i18n.LocaleItIt),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerFrBot", "203397175:AAEqqh2k2QFneWzJ_CmIJ3CHp7cjLa9Pptc", "", "", common.GA_TRACKING_ID, i18n.LocaleFrFr),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerDeBot", "211199220:AAEia3GkoOOX61aygVJdVnxU83PQJpftae4", "", "", common.GA_TRACKING_ID, i18n.LocaleDeDe),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerPLbot", "254844727:AAG3a_1wgSuu77gWmKrcnUy0KN7Yrt0MhO8", "", "", common.GA_TRACKING_ID, i18n.LocalePlPl),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerPtBot", "236826743:AAGx0uDsCO0RZap84IO7dzVSszfA_0HE1m4", "", "", common.GA_TRACKING_ID, i18n.LocalePtBr),
				telegram.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerEsBot", "189365214:AAGnXfb8qqUou__-X5foSGSGfgOkXDm9wV4", "", "", common.GA_TRACKING_ID, i18n.LocalePtBr),
			)
		case strongo.EnvDevTest:
			_bots = botsfw.NewBotSettingsBy(router,
				// Development bots
				telegram.NewTelegramBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev1Bot", "256321815:AAEmCyeWYIIL7TZhJZIqTHohtR3RP7MOOTY", "", "", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				telegram.NewTelegramBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev1RuBot", "395833888:AAF-1QnJvy5tOk4LSfIan07AFuEJcldszhs", "", "", common.GA_TRACKING_ID, i18n.LocaleRuRu),
				//telegram.NewTelegramBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev2RuBot", "360514041:AAFXuT0STHBD9cOn1SFmKzTYDmalP0Rz-7M", "", "", common.GA_TRACKING_ID, i18n.LocalesByCode5[i18n.LocalCodeRuRu]),
			)
		case strongo.EnvStaging:
			_bots = botsfw.NewBotSettingsBy(router,
				// Staging bots
				telegram.NewTelegramBot(strongo.EnvStaging, bot.ProfileDebtus, "DebtsTrackerSt1Bot", "254651741:AAFY_jdNxZHZ5OEIu4VEr5tdcSPSAYnLLWE", "", "", common.GA_TRACKING_ID, i18n.LocaleEnUS),
			)
		case strongo.EnvLocal:
			_bots = botsfw.NewBotSettingsBy(router,
				// Staging bots
				telegram.NewTelegramBot(strongo.EnvLocal, bot.ProfileDebtus, "DebtsTrackerLocalBot", "334671898:AAG38EvZhGb3FTCttyCoSwtmQGFeZ20SqdQ", "", "", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				telegram.NewTelegramBot(strongo.EnvLocal, bot.ProfileSplitus, "SplitusLocalBot", "447286300:AAF6qaS1rp7zfdB3h56lkzrReAHpEWKKYLY", "", "", common.GA_TRACKING_ID, i18n.LocaleEnUS),
				telegram.NewTelegramBot(strongo.EnvLocal, bot.ProfileCollectus, "CollectusLocalBot", "471286497:AAF1m-0jqQeJyXSH1gKMRSeX87Xr_HnIiII", "", "", common.GA_TRACKING_ID, i18n.LocaleEnUS),
			)
		default:
			panic(fmt.Sprintf("Unknown environment => %v:%v", environment, strongo.EnvironmentNames[environment]))
		}
	}
	return _bots
}

func GetBotSettingsByLang(environment strongo.Environment, profile, lang string) (botsfw.BotSettings, error) {
	botSettingsBy := Bots(environment, nil)
	langLen := len(lang)
	if langLen == 2 {
		lang = fmt.Sprintf("%v-%v", strings.ToLower(lang), strings.ToUpper(lang))
	} else if langLen != 5 {
		return botsfw.BotSettings{}, fmt.Errorf("Invalid length of lang parameter: %v, %v", langLen, lang)
	}
	findByProfile := func(botSettings []botsfw.BotSettings) (botsfw.BotSettings, error) {
		for _, bs := range botSettings {
			if bs.Profile == profile {
				return bs, nil
			}
		}
		return botsfw.BotSettings{}, fmt.Errorf("Not found by locale=%v + profile=%v", lang, profile)
	}
	if botSettings, ok := botSettingsBy.ByLocale[lang]; ok {
		return findByProfile(botSettings)
	} else if lang != DEFAULT_LOCALE {
		if botSettings, ok = botSettingsBy.ByLocale[DEFAULT_LOCALE]; ok {
			return findByProfile(botSettings)
		}
	}
	return botsfw.BotSettings{}, fmt.Errorf("no bot setting for both %s & %s locales", lang, DEFAULT_LOCALE)
}
