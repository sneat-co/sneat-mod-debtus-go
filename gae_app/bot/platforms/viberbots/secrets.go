package viberbots

import (
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/bot"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"context"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/platforms/viber"
	"google.golang.org/appengine"
)

var _bots botsfw.SettingsBy

func Bots(c context.Context) botsfw.SettingsBy { //TODO: Consider to do pre-deployment replace
	if len(_bots.ByCode) == 0 {
		host := appengine.DefaultVersionHostname(c)
		if host == "" || strings.Contains(host, "dev") {
			_bots = botsfw.NewBotSettingsBy(nil,
				// Development bot
				viber.NewViberBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev", "451be8dd024fbbc7-4fb4285be8dbb24e-1b2d99610f798855", "", strongo.LocalesByCode5[strongo.LocaleCodeEnUS]),
			)
		} else if strings.Contains(host, "st1") {
			//_bots = botsfw.NewBotSettingsBy(
			//	// Staging bots
			//)
		} else if strings.HasPrefix(host, "debtstracker-io.") {
			_bots = botsfw.NewBotSettingsBy(nil,
				// Production bot
				viber.NewViberBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTracker", "4512c8fee64003e3-c80409381d9f87ff-b0f58459c505b13d", common.GA_TRACKING_ID, strongo.LocalesByCode5[strongo.LocaleCodeEnUS]),
			)
		}
	}
	return _bots
}

// TODO: Decouple to common lib
//func GetBotSettingsByLang(c context.Context, lang string) (bots.BotSettings, error) {
//	botSettingsBy := Bots(c)
//	langLen := len(lang)
//	if langLen == 2 {
//		lang = fmt.Sprintf("%v-%v", strings.ToLower(lang), strings.ToUpper(lang))
//	} else if langLen != 5 {
//		return botsfw.BotSettings{}, fmt.Errorf("Invalid length of lang parameter: %v, %v", langLen, lang)
//	}
//	if botSettings, ok := botSettingsBy.Locale[lang]; ok {
//		return botSettings, nil
//	} else if lang != DEFAULT_LOCALE {
//		if botSettings, ok = botSettingsBy.Locale[DEFAULT_LOCALE]; ok {
//			return botSettings, nil
//		}
//	}
//	return botsfw.BotSettings{}, fmt.Errorf("No bot setting for both %v & %v locales.", lang, DEFAULT_LOCALE)
//}
