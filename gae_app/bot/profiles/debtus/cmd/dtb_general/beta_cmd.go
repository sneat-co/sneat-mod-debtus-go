package dtb_general

import (
	"fmt"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/bots-framework/core"
)

const BETA_COMMAND = "beta"

var BetaCommand = bots.Command{
	Code:     BETA_COMMAND,
	Commands: []string{"/beta"},
	Action: func(whc bots.WebhookContext) (bots.MessageFromBot, error) {
		bot := whc.GetBotSettings()
		token := auth.IssueToken(whc.AppUserIntID(), whc.BotPlatform().ID()+":"+bot.Code, false)
		host := common.GetWebsiteHost(bot.Code)
		betaUrl := fmt.Sprintf(
			"https://%v/app/#lang=%v&secret=%v",
			host, whc.Locale().SiteCode(), token,
		)
		return whc.NewMessage(betaUrl), nil
	},
}
