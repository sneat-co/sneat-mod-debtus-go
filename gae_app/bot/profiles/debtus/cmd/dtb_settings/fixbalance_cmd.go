package dtb_settings

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"context"
	"github.com/crediterra/money"
	"github.com/strongo/bots-framework/core"
)

var FixBalanceCommand = bots.Command{
	Code:     "fixbalance",
	Commands: []string{"/fixbalance"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		if err = dtdal.DB.RunInTransaction(whc.Context(), func(c context.Context) error {
			user, err := facade.User.GetUserByID(c, whc.AppUserIntID())
			if err != nil {
				return err
			}
			contacts := user.Contacts()
			balance := make(money.Balance, user.BalanceCount)
			for _, contact := range contacts {
				b := contact.Balance()
				for k, v := range b {
					balance[k] += v
				}
			}
			user.SetBalance(balance)
			return facade.User.SaveUser(c, user)
		}, dtdal.CrossGroupTransaction); err != nil {
			return
		}
		m = whc.NewMessage("Balance fixed")
		return
	},
}
