package models

import (
	"fmt"
	"github.com/strongo/decimal"
)

type Amount struct {
	Currency Currency
	Value    decimal.Decimal64p2
}

func NewAmount(currency Currency, value decimal.Decimal64p2) Amount {
	if currency == "" {
		panic("Currency not provided")
	}
	return Amount{
		Currency: currency,
		Value:    value,
	}
}

func (a Amount) String() string {
	//if currencySign, ok := currencySigns[a.Currency]; ok {
	//	return fmt.Sprintf("%v%v", currencySign, a.Value)
	//}
	return fmt.Sprintf("%v %v", a.Value, a.Currency)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type amountSorter struct {
	amounts []Amount
}

// Len is part of sort.Interface.
func (s *amountSorter) Len() int {
	return len(s.amounts)
}

// Swap is part of sort.Interface.
func (s *amountSorter) Swap(i, j int) {
	s.amounts[i], s.amounts[j] = s.amounts[j], s.amounts[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *amountSorter) Less(i, j int) bool {
	return s.amounts[i].Value > s.amounts[j].Value // Reverse sort - large amounts first
}
