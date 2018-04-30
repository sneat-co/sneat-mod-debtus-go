package tgbots

import (
	"testing"

	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"github.com/strongo/app"
)

func TestGetBotSettingsByLang(t *testing.T) {
	verify := func(profile, locale, code string) {
		botSettings, err := GetBotSettingsByLang(strongo.EnvProduction, bot.ProfileDebtus, locale)
		if err != nil {
			t.Fatal(err)
		}
		if botSettings.Code != code {
			t.Error(code + " not found in settings, got: " + botSettings.Code)
		}
	}
	verify(bot.ProfileDebtus, strongo.LocalCodeRuRu, "DebtsTrackerRuBot")
	verify(bot.ProfileDebtus, strongo.LocaleCodeEnUS, "DebtsTrackerBot")
}
