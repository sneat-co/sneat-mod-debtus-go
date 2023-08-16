package tgbots

import (
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/bot"
	strongo "github.com/strongo/app"
	"github.com/strongo/i18n"
	"testing"
)

func TestGetBotSettingsByLang(t *testing.T) {
	t.Skip("TODO: fix this test to run on CI")
	verify := func(profile, locale, code string) {
		botSettings, err := GetBotSettingsByLang(strongo.EnvLocal, bot.ProfileDebtus, locale)
		if err != nil {
			t.Fatal(err)
		}
		if botSettings.Code != code {
			t.Error(code + " not found in settings, got: " + botSettings.Code)
		}
	}
	verify(bot.ProfileDebtus, i18n.LocalCodeRuRu, "DebtsTrackerRuBot")
	verify(bot.ProfileDebtus, i18n.LocaleCodeEnUS, "DebtsTrackerBot")
}
