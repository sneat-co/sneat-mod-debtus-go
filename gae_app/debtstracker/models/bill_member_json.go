package models

//go:generate ffjson $GOFILE

import (
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/decimal"
)

type BillMemberJson struct {
	MemberJson
	Owes       decimal.Decimal64p2 `json:",omitempty"`
	Percent    decimal.Decimal64p2 `json:",omitempty"`
	Adjustment decimal.Decimal64p2 `json:",omitempty"`
	Paid       decimal.Decimal64p2 `json:",omitempty"`
	//TransferIDs []int64             `json:",omitempty"`
}

func (m *BillMemberJson) String() string {
	buffer, _ := ffjson.MarshalFast(m)
	return string(buffer)
}

type MemberContactJson struct {
	ContactID   int64
	ContactName string
}

type MemberContactsJsonByUser map[string]MemberContactJson
