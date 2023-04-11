package facade

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/crediterra/money"
	"github.com/strongo/decimal"
)

type SplitMemberTotal struct {
	models.BillMemberBalance
}

func (t SplitMemberTotal) Balance() decimal.Decimal64p2 {
	return t.Paid - t.Owes
}

type SplitBalanceByMember map[string]SplitMemberTotal
type SplitBalanceByCurrencyAndMember map[money.Currency]SplitBalanceByMember

func (billFacade) getBalances(splitID int64, bills []models.Bill) (balanceByCurrency SplitBalanceByCurrencyAndMember) {
	balanceByCurrency = make(SplitBalanceByCurrencyAndMember)
	for _, bill := range bills {
		var (
			balanceByMember SplitBalanceByMember
			ok              bool
		)
		if balanceByMember, ok = balanceByCurrency[bill.Data.Currency]; !ok {
			balanceByMember = make(SplitBalanceByMember)
			balanceByCurrency[bill.Data.Currency] = balanceByMember
		}
		for memberID, memberBalance := range bill.Data.GetBalance() {
			memberTotal := balanceByMember[memberID]
			memberTotal.Paid += memberBalance.Paid
			memberTotal.Owes += memberBalance.Owes
			balanceByMember[memberID] = memberTotal
		}
	}
	return
}

func (billFacade) cleanupBalances(balanceByCurrency SplitBalanceByCurrencyAndMember) SplitBalanceByCurrencyAndMember {
	return balanceByCurrency
}
