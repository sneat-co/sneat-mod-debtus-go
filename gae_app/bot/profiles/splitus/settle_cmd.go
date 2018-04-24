package splitus

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"net/url"
)

const settleBillsCommandCode = "settle"

var settleBillsCommand = bots.Command{
	Code:     settleBillsCommandCode,
	Commands: []string{"/settle"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return settleBillsAction(whc)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		return settleBillsAction(whc)
	},
}

func settleBillsAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "settleBillsAction()")
	var user models.AppUser
	if user, err = dal.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
		return
	}

	outstandingBills := user.GetOutstandingBills()

	m.Text = fmt.Sprintf("len(outstandingBills): %v", len(outstandingBills))

	return
}
