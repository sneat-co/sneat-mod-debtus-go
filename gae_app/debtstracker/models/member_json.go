package models

//go:generate ffjson $GOFILE

import (
	"github.com/strongo/decimal"
	"github.com/pquerna/ffjson/ffjson"
)

type MemberJson struct {
	// We use string IDs as it's faster to marshal and will be more compact in future
	ID            string // required
	Name          string // required
	AddedByUserID string                   `json:",omitempty"`
	UserID        string                   `json:",omitempty"`
	TgUserID      string                   `json:",omitempty"`
	ContactIDs    []string                 `json:",omitempty"`
	ContactByUser MemberContactsJsonByUser `json:",omitempty"`
	Shares        int                      `json:",omitempty"`
}

var _ SplitMember = (*MemberJson)(nil)

func (m MemberJson) GetID() string {
	return m.ID
}

func (m MemberJson) GetName() string {
	return m.Name
}

func (m MemberJson) GetShares() int {
	return m.Shares
}

type GroupMemberJson struct {
	MemberJson
	Balance Balance `json:",omitempty"`
}

var _ SplitMember = (*GroupMemberJson)(nil)

func (m *GroupMemberJson) String() string {
	buffer, _ := ffjson.MarshalFast(m)
	return string(buffer)
}

type GroupBalanceByCurrencyAndMember map[Currency]map[string]decimal.Decimal64p2

type BillMemberJson struct {
	MemberJson
	Paid       decimal.Decimal64p2 `json:",omitempty"`
	Owes       decimal.Decimal64p2 `json:",omitempty"`
	Percent    decimal.Decimal64p2 `json:",omitempty"`
	Adjustment decimal.Decimal64p2 `json:",omitempty"`
	//TransferIDs []int64             `json:",omitempty"`
}

func (m *BillMemberJson) String() string {
	buffer, _ := ffjson.MarshalFast(m)
	return string(buffer)
}

type MemberContactJson struct {
	ContactID   string
	ContactName string
}

type MemberContactsJsonByUser map[string]MemberContactJson
