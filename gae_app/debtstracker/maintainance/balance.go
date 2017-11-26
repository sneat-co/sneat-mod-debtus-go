package maintainance

import (
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
)

func FixBalanceCurrencies(balance models.Balance) (changed bool) {
	euro := models.Currency("euro")
	for c, v := range balance {
		if c == euro {
			c = models.CURRENCY_EUR
		} else if len(c) == 3 {
			cc := strings.ToUpper(string(c))
			if cc != string(c) {
				if cu := models.Currency(cc); cu.IsMoney() {
					balance[cu] += v
					delete(balance, c)
					changed = true
				}
			}
		}
	}
	return
}
