package facade

import (
	"github.com/strongo/decimal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
)

type SplitMemberTotal struct {
	models.BillMemberBalance
}

func (t SplitMemberTotal) Balance() decimal.Decimal64p2 {
	return t.Paid - t.Owes
}

type SplitBalanceByMember map[string]SplitMemberTotal
type SplitBalanceByCurrencyAndMember map[string]SplitBalanceByMember

func (billFacade) getBalances(splitID int64, bills []models.Bill) (balanceByCurrency SplitBalanceByCurrencyAndMember) {
	balanceByCurrency = make(SplitBalanceByCurrencyAndMember)
	for _, bill := range bills {
		for currency, billBalanceByCurrencyAndMember := range bill.GetBalance() {
			var balanceByMember SplitBalanceByMember
			var ok bool

			if balanceByMember, ok = balanceByCurrency[currency]; !ok {
				balanceByMember = make(SplitBalanceByMember)
				balanceByCurrency[currency] = balanceByMember
			}

			for memberID, memberBalance := range billBalanceByCurrencyAndMember {
				memberTotal := balanceByMember[memberID]
				memberTotal.Paid += memberBalance.Paid
				memberTotal.Owes += memberBalance.Owes
				balanceByMember[memberID] = memberTotal
			}
		}
	}
	return
}

func (billFacade) cleanupBalances(balanceByCurrency SplitBalanceByCurrencyAndMember) SplitBalanceByCurrencyAndMember {
	return balanceByCurrency
}

