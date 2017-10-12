package models

import (
	"testing"
	"github.com/strongo/decimal"
)

func TestUpdateMemberOwesForEqualSplit(t *testing.T) {
	var members []BillMemberJson

	members = []BillMemberJson{{}, {}, {}, {}}
	updateMemberOwesForEqualSplit(1000, -1, members)
	t.Logf("members +v: %+v", members)
}

func validateTotal(t *testing.T, members []BillMemberJson, expectedTotal decimal.Decimal64p2) {
	var total decimal.Decimal64p2
	for _, m := range members {
		total += m.Owes
	}
	if total != expectedTotal {
		t.Fatal("Wrong total", total)
	}
}

func TestUpdateMemberOwesForSplitByShares(t *testing.T) {
	var members []BillMemberJson


	members = []BillMemberJson{{}, {}, {}, {}}
	updateMemberOwesForSplitByShares(1000, 0, members)
	t.Logf("members +v: %+v", members)
	if members[0].Owes != 250 || members[1].Owes != 250 || members[2].Owes != 250 || members[3].Owes != 250 {
		t.Error(members)
		return
	}
	validateTotal(t, members, 1000)

	members = []BillMemberJson{{}, {}, {}}
	updateMemberOwesForSplitByShares(1000, -1, members)
	t.Logf("members +v: %+v", members)
	if members[0].Owes != 334 || members[1].Owes != 333 || members[2].Owes != 333 {
		t.Error(members)
		return
	}
	validateTotal(t, members, 1000)

	members = []BillMemberJson{{MemberJson: MemberJson{Shares: 3}}, {MemberJson: MemberJson{Shares: 2}}, {MemberJson: MemberJson{Shares: 1}}}
	updateMemberOwesForSplitByShares(1000, -1, members)
	t.Logf("members +v: %+v", members)
	validateTotal(t, members, 1000)
}
