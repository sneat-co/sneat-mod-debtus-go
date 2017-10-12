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
	Currency     Currency            `json:"c"`
}

type BillMemberBalance struct {
	Paid decimal.Decimal64p2
	Owes decimal.Decimal64p2
}

func (t BillMemberBalance) Balance() decimal.Decimal64p2 {
	return t.Paid - t.Owes
}

type BillBalanceByMember map[string]BillMemberBalance

type BillBalanceDifference BillBalanceByMember

func (diff BillBalanceDifference) IsAffectingGroupBalance() bool {
	for _, m := range diff {
		if m.Paid != m.Owes {
			return true
		}
	}
	return false
}

func (current BillBalanceByMember) BillBalanceDifference(previous BillBalanceByMember) (difference BillBalanceDifference) {
	capacity := len(current) + 1
	if len(previous) > capacity {
		capacity = len(previous) + 1
	}
	difference = make(BillBalanceDifference, capacity)

	for memberID, mCurrent := range current {
		mPrevious := previous[memberID]
		diff := BillMemberBalance{
			Paid: mCurrent.Paid - mPrevious.Paid,
			Owes: mCurrent.Owes - mPrevious.Owes,
		}
		if diff.Paid != 0 || diff.Owes != 0 {
			difference[memberID] = diff
		}
	}

	for memberID, mPrevious := range previous {
		if _, ok := current[memberID]; !ok {
			difference[memberID] = BillMemberBalance{
				Paid: -mPrevious.Paid,
				Owes: -mPrevious.Owes,
			}
		}
	}

	return
}
