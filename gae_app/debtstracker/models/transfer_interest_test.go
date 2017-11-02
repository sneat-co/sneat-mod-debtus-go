package models

import (
	"testing"
	"github.com/strongo/decimal"
	"time"
)

func TestTransferEntityGetOutstandingAmount(t *testing.T) {
	now := time.Now()
	transfer := &TransferEntity{
		DtCreated: now.Add(-3*24*time.Hour - time.Hour/2),
		AmountInCents: decimal.NewDecimal64p2(100, 50),
		TransferInterest: TransferInterest{
			InterestType: InterestPercentSimple,
			InterestPeriod: InterestRatePeriod(7),
			InterestPercent: 7,
		},
	}
	if v := transfer.GetOutstandingAmount().Value; v != decimal.NewDecimal64p2(103, 52) {
		t.Errorf("Expected 103, got: %v", v)
	}
}
