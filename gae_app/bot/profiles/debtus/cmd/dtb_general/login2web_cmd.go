package dtb_general

import (
	"fmt"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
)

const LOGIN2WEB_COMMAND = "login2web"

var Login2WebCommand = bots.Command{
	Code:     LOGIN2WEB_COMMAND,
	Commands: []string{"/login"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		mt := whc.Translate(trans.MESSAGE_TEXT_LOGIN_TO_WEB_APP)
		linker := common.NewLinkerFromWhc(whc)
		mt = strings.Replace(mt, "<a>", fmt.Sprintf(`<a href="%v">`, linker.ToMainScreen(whc)), 1)
		m = whc.NewMessage(mt)
		m.Format = bots.MessageFormatHTML
		m.DisableWebPagePreview = true
		return
	},
}
