package dtb_general

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/auth"
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"fmt"
	"github.com/strongo/bots-framework/core"
)

const BETA_COMMAND = "beta"

var BetaCommand = bots.Command{
	Code:     BETA_COMMAND,
	Commands: []string{"/beta"},
	Action: func(whc bots.WebhookContext) (bots.MessageFromBot, error) {
		bot := whc.GetBotSettings()
		token := auth.IssueToken(whc.AppUserIntID(), whc.BotPlatform().Id()+":"+bot.Code, false)
		host := common.GetWebsiteHost(bot.Code)
		betaUrl := fmt.Sprintf(
			"https://%v/app/#lang=%v&secret=%v",
			host, whc.Locale().SiteCode(), token,
		)
		return whc.NewMessage(betaUrl), nil
	},
}
