package viber

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/viber"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"strings"
)

var _bots bots.SettingsBy

func Bots(c context.Context) bots.SettingsBy { //TODO: Consider to do pre-deployment replace
	if len(_bots.ByCode) == 0 {
		host := appengine.DefaultVersionHostname(c)
		if strings.Contains(host, "dev") {
			_bots = bots.NewBotSettingsBy(nil,
				// Development bot
				viber_bot.NewViberBot(strongo.EnvDevTest, bot.ProfileDebtus, "DebtsTrackerDev", "451be8dd024fbbc7-4fb4285be8dbb24e-1b2d99610f798855", strongo.LocalesByCode5[strongo.LOCALE_EN_US]),
			)
		} else if strings.Contains(host, "st1") {
			//_bots = bots.NewBotSettingsBy(
			//	// Staging bots
			//)
		} else if strings.HasPrefix(host, "debtstracker-io.") {
			_bots = bots.NewBotSettingsBy(nil,
				// Production bot
				viber_bot.NewViberBot(strongo.EnvProduction, bot.ProfileDebtus, "DebtsTracker", "4512c8fee64003e3-c80409381d9f87ff-b0f58459c505b13d", strongo.LocalesByCode5[strongo.LOCALE_EN_US]),
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
//		return bots.BotSettings{}, errors.New(fmt.Sprintf("Invalid length of lang parameter: %v, %v", langLen, lang))
//	}
//	if botSettings, ok := botSettingsBy.Locale[lang]; ok {
//		return botSettings, nil
//	} else if lang != DEFAULT_LOCALE {
//		if botSettings, ok = botSettingsBy.Locale[DEFAULT_LOCALE]; ok {
//			return botSettings, nil
//		}
//	}
//	return bots.BotSettings{}, errors.New(fmt.Sprintf("No bot setting for both %v & %v locales.", lang, DEFAULT_LOCALE))
//}
