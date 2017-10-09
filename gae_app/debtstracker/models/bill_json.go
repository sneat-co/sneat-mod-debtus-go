package models

//go:generate ffjson $GOFILE

import "github.com/strongo/decimal"

type BillJson struct {
	ID           string
	Name         string              `json:"n"`
	MembersCount int                 `json:"m"`
	Total        decimal.Decimal64p2 `json:"t"`
	Currency     string              `json:"c"`
}
