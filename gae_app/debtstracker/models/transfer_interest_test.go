package models

import (
	"testing"
	"github.com/strongo/decimal"
	"time"
)

func TestTransferEntityGetOutstandingAmount(t *testing.T) {
	now := time.Now()
	const day = 24 * time.Hour
	transfer := &TransferEntity{
		DtCreated:     now.Add(-3 * day),
		AmountInCents: decimal.NewDecimal64p2(100, 0),
		TransferInterest: TransferInterest{
			InterestType:    InterestPercentSimple,
			InterestPeriod:  InterestRatePeriod(7),
			InterestPercent: 7,
		},
	}
	checkInterestAndOutstanding := func (expected decimal.Decimal64p2) bool {
		t.Helper()
		if v := transfer.GetOutstandingAmount().Value; v != expected {
			t.Errorf("Expected %v, got: %v", expected, v)
			return false
		}
		return true
	}

	if !checkInterestAndOutstanding(decimal.NewDecimal64p2(104, 0)) {
		return
	}

	transfer.SetReturns([]TransferReturnJson{
		{
			Time: now.Add(-2 * day),
			Amount: decimal.NewDecimal64p2(50, 0),
		},
	})
	if !checkInterestAndOutstanding(decimal.NewDecimal64p2(103, 0)) {
		return
	}
}
