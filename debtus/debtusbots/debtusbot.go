package debtusbots

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/debtusbots/platforms/debtustgbots"
)

var profile botsfw.BotProfile

func GetProfile(errFooterText func() string) botsfw.BotProfile {
	if profile == nil {
		profile = debtustgbots.GetDebtusBotProfile(errFooterText)
	}
	return profile
}
