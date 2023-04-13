package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/log"
	"net/url"
)

const settleBillsCommandCode = "settle"

var settleBillsCommand = botsfw.Command{
	Code:     settleBillsCommandCode,
	Commands: []string{"/settle"},
	Action: func(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
		return settleBillsAction(whc)
	},
	CallbackAction: func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
		return settleBillsAction(whc)
	},
}

func settleBillsAction(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "settleBillsAction()")
	var user models.AppUser
	if user, err = facade.User.GetUserByID(c, nil, whc.AppUserIntID()); err != nil {
		return
	}

	outstandingBills := user.Data.GetOutstandingBills()

	m.Text = fmt.Sprintf("len(outstandingBills): %v", len(outstandingBills))

	return
}
