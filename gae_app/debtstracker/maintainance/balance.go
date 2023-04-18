package maintainance

import (
	"github.com/crediterra/money"
	"strings"
)

func FixBalanceCurrencies(balance money.Balance) (changed bool) {
	euro := money.Currency("euro")
	for c, v := range balance {
		if c == euro {
			c = money.CURRENCY_EUR
		}
		if len(c) == 3 {
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
