package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/decimal"
)

var simpleFor7daysAt7percent = TransferInterest{
	InterestType:    InterestPercentSimple,
	InterestPeriod:  7,
	InterestPercent: decimal.NewDecimal64p2(7, 0),
	//InterestMinimumPeriod: 5,
}

const day = 24 * time.Hour

func assertOutstandingValue(t *testing.T, transfer *TransferEntity, expected decimal.Decimal64p2) bool {
	t.Helper()
	if v := transfer.GetOutstandingValue(time.Now()); v != expected {
		t.Errorf("Expected %v, got: %v", expected, v)
		return false
	}
	return true
}

func TestTransferEntity_GetInterestValue(t *testing.T) {
	now := time.Now()
	transfer := &TransferEntity{
		DtCreated:     now,
		AmountInCents: decimal.NewDecimal64p2(10, 0),
		TransferInterest: TransferInterest{
			InterestType:          InterestPercentSimple,
			InterestPeriod:        3,
			InterestPercent:       decimal.NewDecimal64p2(3, 0),
			InterestMinimumPeriod: 3,
		},
	}

	if !assertOutstandingValue(t, transfer, decimal.NewDecimal64p2(10, 30)) {
		return
	}

	transfer.SetReturns([]TransferReturnJson{
		{
			Time:   now,
			Amount: decimal.NewDecimal64p2(10, 30),
		},
	})
	if !assertOutstandingValue(t, transfer, 0) {
		return
	}
}

func TestTransferEntityGetOutstandingAmount(t *testing.T) {
	now := time.Now()
	transfer := &TransferEntity{
		DtCreated:        now.Add(-3 * day),
		AmountInCents:    decimal.NewDecimal64p2(100, 0),
		TransferInterest: simpleFor7daysAt7percent,
	}

	if !assertOutstandingValue(t, transfer, decimal.NewDecimal64p2(104, 0)) {
		return
	}

	transfer.SetReturns([]TransferReturnJson{
		{
			Time:   now.Add(-2 * day),
			Amount: decimal.NewDecimal64p2(50, 0),
		},
	})
	if !assertOutstandingValue(t, transfer, decimal.NewDecimal64p2(103, 0)) {
		return
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
	balanceWithInterest := userContact.BalanceWithInterest(nil, time.Now())
	if len(balanceWithInterest) != 1 {
		t.Fatalf("len(balanceWithInterest) != 1: %v", len(balanceWithInterest))
	}
	if expected := decimal.NewDecimal64p2(100+4, 0); balanceWithInterest["EUR"] != expected {
		t.Errorf("Expected %v, got %v", expected, balanceWithInterest["EUR"])
	}
}

func Test_updateBalanceWithInterest(t *testing.T) {
	balance := Balance{
		CURRENCY_EUR: decimal.NewDecimal64p2FromFloat64(52.00),
	}
	now := time.Now()
	outstandingWithInterest := []TransferWithInterestJson{
		{
			TransferInterest: TransferInterest{
				InterestType:          InterestPercentSimple,
				InterestPercent:       decimal.NewDecimal64p2FromFloat64(2.00),
				InterestPeriod:        1,
				InterestMinimumPeriod: 1,
			},
			Starts:   now,
			Currency: CURRENCY_EUR,
			Amount:   decimal.NewDecimal64p2FromFloat64(100.00),
			Returns: []TransferReturnJson{
				{
					Time:   now.Add(time.Minute),
					Amount: decimal.NewDecimal64p2FromFloat64(50.00),
				},
			},
		},
	}
	updateBalanceWithInterest(nil, balance, outstandingWithInterest, now.Add(time.Hour))
	t.Log(balance)
}
