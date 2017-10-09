package telegram

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot"
	"github.com/strongo/app"
	"testing"
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
	verify(bot.ProfileDebtus, strongo.LOCALE_RU_RU, "DebtsTrackerRuBot")
	verify(bot.ProfileDebtus, strongo.LOCALE_EN_US, "DebtsTrackerBot")
}
