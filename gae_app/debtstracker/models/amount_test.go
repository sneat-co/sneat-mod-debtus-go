package models

import (
	"testing"

	"github.com/strongo/decimal"
)

func TestNewAmount(t *testing.T) {
	rub := Currency(CURRENCY_RUB)
	amount := NewAmount(rub, decimal.NewDecimal64p2FromFloat64(123.45))
	if amount.Currency != rub {
		t.Error("amount.Currency != rub")
	}
	if amount.Value != decimal.NewDecimal64p2FromFloat64(123.45) {
		t.Error("amount.Value != 123.45")
	}
}
