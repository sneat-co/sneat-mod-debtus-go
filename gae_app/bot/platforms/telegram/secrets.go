package telegram

import (
	"fmt"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
)

var _bots bots.SettingsBy

const DEFAULT_LOCALE = strongo.LOCALE_EN_US

const DebtusBotToken = "467112035:AAG9Hij0ofnI6GGXyuc6zol0F4XGQ4OK5Tk"

func Bots(environment strongo.Environment, router func(profile string) bots.WebhooksRouter) bots.SettingsBy { //TODO: Consider to do pre-deployment replace
	if len(_bots.ByCode) == 0 || (!_bots.HasRouter && router != nil) {
		//log.Debugf(c, "Bots() => hostname:%v, environment:%v:%v", hostname, environment, strongo.EnvironmentNames[environment])
		switch environment {
		case strongo.EnvProduction:
			_bots = bots.NewBotSettingsBy(router,
				// Production bots
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileCollectus, "CollectusBot", "458860316:AAFk_hOXK5vFWu43jp4apWgQjmHHv87CU9E", "", "", strongo.LocaleEnUS),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileSplitus, "SplitusBot", "345328965:AAHmM7rUCwiPBlVIv-IfhrWhYIUVSHerkpg", "", "", strongo.LocaleEnUS),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerBot", "209808950:AAHEwdBVtVIhKZhieTCP6zdbVkTROoj0fyA", "284685063:TEST:Njc4MWQ2NzlmMDAx", "350862534:LIVE:ZjAzOWE3ODg5OWMy", strongo.LocaleEnUS),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerRuBot", "218446201:AAGyvWHuodNYT8kgbR_701m6y8Xg5D9iTSA", "284685063:TEST:MDg3NzM5ZTUxMTNk	", "350862534:LIVE:MGM1ODY0N2Q2ZDM5", strongo.LocaleRuRu),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerFaBot", "182148042:AAFHD7MfWr5CLjGczaiqsx-Oo6msoR_5JfM", "", "", strongo.LocalesByCode5[strongo.LOCALE_FA_IR]),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerItBot", "143800015:AAFrLrjyKCIqVFE0YsdZghYtDVmiLpa_P_A", "84685063:TEST:Zjg1ZTIxYzEyNTQ3", "350862534:LIVE:ZmRhMWRhOWZiOWIx", strongo.LocaleItIt),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerFrBot", "203397175:AAEqqh2k2QFneWzJ_CmIJ3CHp7cjLa9Pptc", "", "", strongo.LocaleFrFr),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerDeBot", "211199220:AAEia3GkoOOX61aygVJdVnxU83PQJpftae4", "", "", strongo.LocaleDeDe),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerPLbot", "254844727:AAG3a_1wgSuu77gWmKrcnUy0KN7Yrt0MhO8", "", "", strongo.LocalePlPl),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerPtBot", "236826743:AAGx0uDsCO0RZap84IO7dzVSszfA_0HE1m4", "", "", strongo.LocalePtBr),
				telegram_bot.NewTelegramBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTrackerEsBot", "189365214:AAGnXfb8qqUou__-X5foSGSGfgOkXDm9wV4", "", "", strongo.LocalePtBr),

			)
		case strongo.EnvDevTest:
			_bots = bots.NewBotSettingsBy(router,
				// Development bots
				telegram_bot.NewTelegramBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev1Bot", "256321815:AAEmCyeWYIIL7TZhJZIqTHohtR3RP7MOOTY", "", "", strongo.LocaleEnUS),
				telegram_bot.NewTelegramBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev1RuBot", "395833888:AAF-1QnJvy5tOk4LSfIan07AFuEJcldszhs", "", "", strongo.LocaleRuRu),
				//telegram_bot.NewTelegramBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev2RuBot", "360514041:AAFXuT0STHBD9cOn1SFmKzTYDmalP0Rz-7M", "", "", strongo.LocalesByCode5[strongo.LOCALE_RU_RU]),
			)
		case strongo.EnvStaging:
			_bots = bots.NewBotSettingsBy(router,
				// Staging bots
				telegram_bot.NewTelegramBot(strongo.EnvStaging, bot.ProfileDebtus, "DebtsTrackerSt1Bot", "254651741:AAFY_jdNxZHZ5OEIu4VEr5tdcSPSAYnLLWE", "", "", strongo.LocaleEnUS),
			)
		case strongo.EnvLocal:
			_bots = bots.NewBotSettingsBy(router,
				// Staging bots
				telegram_bot.NewTelegramBot(strongo.EnvLocal, bot.ProfileDebtus, "DebtsTrackerLocalBot", "334671898:AAG38EvZhGb3FTCttyCoSwtmQGFeZ20SqdQ", "", "", strongo.LocaleEnUS),
				telegram_bot.NewTelegramBot(strongo.EnvLocal, bot.ProfileSplitus, "SplitusLocalBot", "447286300:AAF6qaS1rp7zfdB3h56lkzrReAHpEWKKYLY", "", "", strongo.LocaleEnUS),
				telegram_bot.NewTelegramBot(strongo.EnvLocal, bot.ProfileCollectus, "CollectusLocalBot", "471286497:AAF1m-0jqQeJyXSH1gKMRSeX87Xr_HnIiII", "", "", strongo.LocaleEnUS),
			)
		default:
			panic(fmt.Sprintf("Unknown environment => %v:%v", environment, strongo.EnvironmentNames[environment]))
		}
	}
	return _bots
}

func GetBotSettingsByLang(environment strongo.Environment, profile, lang string) (bots.BotSettings, error) {
	botSettingsBy := Bots(environment, nil)
	langLen := len(lang)
	if langLen == 2 {
		lang = fmt.Sprintf("%v-%v", strings.ToLower(lang), strings.ToUpper(lang))
	} else if langLen != 5 {
		return bots.BotSettings{}, fmt.Errorf("Invalid length of lang parameter: %v, %v", langLen, lang)
	}
	findByProfile := func(botSettings []bots.BotSettings) (bots.BotSettings, error) {
		for _, bs := range botSettings {
			if bs.Profile == profile {
				return bs, nil
			}
		}
		return bots.BotSettings{}, fmt.Errorf("Not found by locale=%v + profile=%v", lang, profile)
	}
	if botSettings, ok := botSettingsBy.ByLocale[lang]; ok {
		return findByProfile(botSettings)
	} else if lang != DEFAULT_LOCALE {
		if botSettings, ok = botSettingsBy.ByLocale[DEFAULT_LOCALE]; ok {
			return findByProfile(botSettings)
		}
	}
	return bots.BotSettings{}, fmt.Errorf("No bot setting for both %v & %v locales.", lang, DEFAULT_LOCALE)
}
