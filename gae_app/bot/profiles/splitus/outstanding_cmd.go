package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

const outstandingBalanceCommandCode = "outstanding-balance"

var outstandingBalanceCommand = bots.Command{
	Code:     outstandingBalanceCommandCode,
	Commands: []string{"/outstanding"},
	Action:   outstandingBalanceAction,
}

func outstandingBalanceAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "outstandingBalanceAction()")
	var user models.AppUser
	if user, err = facade.User.GetUserByID(c, whc.AppUserIntID()); err != nil {
		return
	}

	outstandingBalance := user.GetOutstandingBalance()
	m.Text = fmt.Sprintf("Outstanding balance: %v", outstandingBalance)
	return
}
