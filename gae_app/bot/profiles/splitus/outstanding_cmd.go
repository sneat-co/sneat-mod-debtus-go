package splitus

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/log"
)

const outstandingBalanceCommandCode = "outstanding-balance"

var outstandingBalanceCommand = botsfw.Command{
	Code:     outstandingBalanceCommandCode,
	Commands: []string{"/outstanding"},
	Action:   outstandingBalanceAction,
}

func outstandingBalanceAction(whc botsfw.WebhookContext) (m botsfw.MessageFromBot, err error) {
	c := whc.Context()
	log.Debugf(c, "outstandingBalanceAction()")
	var user models.AppUser
	if user, err = facade.User.GetUserByID(c, nil, whc.AppUserIntID()); err != nil {
		return
	}

	outstandingBalance := user.Data.GetOutstandingBalance()
	m.Text = fmt.Sprintf("Outstanding balance: %v", outstandingBalance)
	return
}
