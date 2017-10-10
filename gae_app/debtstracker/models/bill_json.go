package models

//go:generate ffjson $GOFILE

import (
	"github.com/strongo/decimal"
)

type BillJson struct {
	ID           string
	Name         string              `json:"n"`
	MembersCount int                 `json:"m"`
	Total        decimal.Decimal64p2 `json:"t"`
	Currency     string              `json:"c"`
}

type BillMemberBalance struct {
	Paid    decimal.Decimal64p2
	Owes    decimal.Decimal64p2
}

func (t BillMemberBalance) Balance() decimal.Decimal64p2 {
	return t.Paid - t.Owes
}

type BillBalanceByMember map[string]BillMemberBalance
type BillBalanceByCurrencyAndMember map[string]BillBalanceByMember
