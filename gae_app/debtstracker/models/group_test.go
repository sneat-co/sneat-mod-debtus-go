package models

import "testing"

func TestGroupEntity_ApplyBillBalanceDifference(t *testing.T) {
	groupEntity := GroupEntity{}

	{  // Try to apply empty difference
		if changed, err := groupEntity.ApplyBillBalanceDifference(BillBalanceDifference{}, "EUR"); err != nil {
			t.Error(err)
		} else if changed {
			t.Error("Should not return changed=true")
		}
	}

	groupEntity.SetGroupMembers([]GroupMemberJson{
		{MemberJson: MemberJson{ID: "m1", UserID: 1}},
	})

	{  // Try to apply difference to empty balance
		if changed, err := groupEntity.ApplyBillBalanceDifference(BillBalanceDifference{"m1": BillMemberBalance{Paid: 1000, Owes: 1000}}, "EUR"); err != nil {
			t.Error(err)
		} else if changed {
			t.Error("Should return changed=false: " + groupEntity.MembersJson)
		}
	}

	members := append(groupEntity.GetGroupMembers(), GroupMemberJson{MemberJson: MemberJson{ID: "m2", UserID: 2}})
	groupEntity.SetGroupMembers(members)


	{  // Try to add another member
		if changed, err := groupEntity.ApplyBillBalanceDifference(BillBalanceDifference{
			"m1": BillMemberBalance{Paid: 0, Owes: -400},
			"m2": BillMemberBalance{Paid: 0, Owes: 400},
			}, "EUR"); err != nil {
			t.Error(err)
		} else if !changed {
			t.Error("Should return changed=true")
		}
	}
}
