package inspector

import (
	"sync"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/decimal"
)

type balanceRow struct {
	// TODO: rename
	user                  decimal.Decimal64p2
	contacts              decimal.Decimal64p2
	transfers             decimal.Decimal64p2
	userContactBalanceErr error
	contactBalanceErr     error
}

type balancesByCurrency struct {
	*sync.Mutex
	err        error
	byCurrency map[models.Currency]balanceRow
}

type balances struct {
	withInterest    balancesByCurrency
	withoutInterest balancesByCurrency
}

func newBalances(who string, withoutInterest, withInterest models.Balance) balances {
	return balances{
		withoutInterest: newBalanceSummary(who, withoutInterest),
		withInterest:    newBalanceSummary(who, withInterest),
	}
}
