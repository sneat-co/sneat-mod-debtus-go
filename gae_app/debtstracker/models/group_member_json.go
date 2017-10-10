package models

//go:generate ffjson $GOFILE

import (
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/decimal"
)

type GroupMemberJson struct {
	MemberJson
	Balance Balance
}

var _ SplitMember = (*GroupMemberJson)(nil)

func (m *GroupMemberJson) String() string {
	buffer, _ := ffjson.MarshalFast(m)
	return string(buffer)
}

type GroupBalanceByCurrencyAndMember map[string]map[string]decimal.Decimal64p2