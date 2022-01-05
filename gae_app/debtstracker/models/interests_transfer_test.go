package models

import (
	"encoding/json"
	"github.com/crediterra/money"
	"testing"
	"time"

	"github.com/crediterra/go-interest"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
)

var simpleFor7daysAt7percent = TransferInterest{
	InterestType:    interest.FormulaSimple,
	InterestPeriod:  7,
	InterestPercent: decimal.NewDecimal64p2(7, 0),
	// InterestMinimumPeriod: 5,
}

const day = 24 * time.Hour

func assertOutstandingValue(t *testing.T, transfer interface {
	GetOutstandingValue(time.Time) decimal.Decimal64p2
}, periodEnds time.Time, expected decimal.Decimal64p2) bool {
	t.Helper()
	if v := transfer.GetOutstandingValue(periodEnds); v != expected {
		t.Errorf("Expected outstanding value to be %v, got: %v", expected, v)
		return false
	}
	return true
}

func TestTransferEntity_GetInterestValue(t *testing.T) {
	now := time.Now()
	transfer := Transfer{
		IntegerID: db.NewIntID(111),
		TransferEntity: &TransferEntity{
			DtCreated:        now,
			IsOutstanding:    true,
			AmountInCents:    1000,
			TransferInterest: NewInterest(interest.FormulaSimple, decimal.FromInt(3), 3).WithMinimumPeriod(3),
		},
	}

	if !assertOutstandingValue(t, transfer.TransferEntity, now, 1030) {
		return
	}

	if err := transfer.AddReturn(TransferReturnJson{TransferID: 123, Time: now, Amount: 1030}); err != nil {
		t.Fatal(err)
	}

	if !assertOutstandingValue(t, transfer.TransferEntity, now, 0) {
		return
	}

	if transfer.IsOutstanding {
		t.Error("Transfer should be NOT outstaning")
	}
}

func TestTransferEntityGetOutstandingValue(t *testing.T) {
	now := time.Now()
	transfer := Transfer{
		IntegerID: db.NewIntID(111),
		TransferEntity: &TransferEntity{
			IsOutstanding:    true,
			DtCreated:        now.Add(-3*day + time.Hour),
			AmountInCents:    decimal.NewDecimal64p2(100, 0),
			TransferInterest: simpleFor7daysAt7percent,
		},
	}

	if !assertOutstandingValue(t, transfer, now, decimal.NewDecimal64p2(103, 0)) {
		return
	}

	if err := transfer.AddReturn(TransferReturnJson{TransferID: 123, Time: transfer.DtCreated.Add(23 * time.Hour), Amount: 3100}); err != nil {
		t.Fatal(err)
	}
	if !transfer.IsOutstanding {
		t.Fatal("transfer should remain outstanding after partial return")
	}
	// expecting 100 + 1 - 31 => 70 + 0.7*2 => 71.40
	if !assertOutstandingValue(t, transfer, now, 7140) {
		return
	}
	if err := transfer.AddReturn(TransferReturnJson{TransferID: 124, Time: now, Amount: 7140}); err != nil {
		t.Fatal(err)
	}
	if transfer.IsOutstanding {
		t.Fatal("transfer should be NOT outstanding after full return")
	}
}

func TestUserContactJson_BalanceWithInterest(t *testing.T) {
	now := time.Now()

	balanceJson := json.RawMessage(`{"EUR": 100}`)
	userContact := UserContactJson{
		BalanceJson: &balanceJson,
		Transfers: &UserContactTransfersInfo{
			OutstandingWithInterest: []TransferWithInterestJson{
				{
					TransferID:       1,
					Currency:         "EUR",
					Amount:           decimal.NewDecimal64p2(100, 0),
					Starts:           now.Add(-3 * day),
					TransferInterest: simpleFor7daysAt7percent,
				},
			},
		},
	}
	data, err := ffjson.Marshal(userContact)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
	err = ffjson.Unmarshal(data, &userContact)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", userContact.Transfers.OutstandingWithInterest[0])
	balanceWithInterest, _ := userContact.BalanceWithInterest(nil, time.Now())
	if len(balanceWithInterest) != 1 {
		t.Fatalf("len(balanceWithInterest) != 1: %v", len(balanceWithInterest))
	}
	if expected := decimal.NewDecimal64p2(100+4, 0); balanceWithInterest["EUR"] != expected {
		t.Errorf("Expected %v, got %v", expected, balanceWithInterest["EUR"])
	}
}

func Test_updateBalanceWithInterest(t *testing.T) {
	balance := money.Balance{
		money.CURRENCY_EUR: decimal.NewDecimal64p2FromFloat64(52.00),
	}
	now := time.Now()
	outstandingWithInterest := []TransferWithInterestJson{
		{
			TransferInterest: NewInterest(interest.FormulaSimple, decimal.FromInt(2), 1).WithMinimumPeriod(1),
			Starts:           now,
			Currency:         money.CURRENCY_EUR,
			Amount:           decimal.NewDecimal64p2FromFloat64(100.00),
			Returns: []TransferReturnJson{
				{
					Time:   now.Add(time.Minute),
					Amount: decimal.NewDecimal64p2FromFloat64(50.00),
				},
			},
		},
	}
	if err := updateBalanceWithInterest(true, balance, outstandingWithInterest, now.Add(time.Hour)); err != nil {
		t.Error(err)
	}
	t.Log(balance)
}
