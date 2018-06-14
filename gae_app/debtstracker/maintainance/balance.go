package maintainance

import (
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
)

func FixBalanceCurrencies(balance money.Balance) (changed bool) {
	euro := money.Currency("euro")
	for c, v := range balance {
		if c == euro {
			c = money.Currency_EUR
		} else if len(c) == 3 {
			cc := strings.ToUpper(string(c))
			if cc != string(c) {
				if cu := money.Currency(cc); cu.IsMoney() {
					balance[cu] += v
					delete(balance, c)
					changed = true
				}
			}
		}
	}
	return
}
