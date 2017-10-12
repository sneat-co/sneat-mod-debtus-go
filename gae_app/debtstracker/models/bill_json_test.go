package models

import "testing"

func TestBillBalanceByMember_BillDifference(t *testing.T) {
	previous := BillBalanceByMember{}
	current := BillBalanceByMember{}

	{  // Test empty
		if diff := current.BillBalanceDifference(previous); len(diff) != 0 {
			t.Error("Should be no difference", diff)
		}
	}

	{  // Test non empty current and empty previous
		previous = BillBalanceByMember{}
		current = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 4},
		}
		if diff := current.BillBalanceDifference(previous); len(diff) != 1 {
			t.Error("Should have single item", diff)
		} else if md, ok := diff["m1"]; !ok {
			t.Error("Item should be m1", diff)
		} else if md.Paid != 12 || md.Owes != 4 {
			t.Error("Wrong value", diff)
		}
	}

	{  // Test increase in Paid
		previous = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 10, Owes: 4},
		}
		current = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 4},
		}
		if diff := current.BillBalanceDifference(previous); len(diff) != 1 {
			t.Error("Should have single item", diff)
		} else if md, ok := diff["m1"]; !ok {
			t.Error("Item should be m1", diff)
		} else if md.Paid != 2 || md.Owes != 0 {
			t.Error("Wrong value", diff)
		}
	}

	{  // Test increase in Owes
		previous = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 1},
		}
		current = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 4},
		}
		if diff := current.BillBalanceDifference(previous); len(diff) != 1 {
			t.Error("Should have single item", diff)
		} else if md, ok := diff["m1"]; !ok {
			t.Error("Item should be m1", diff)
		} else if md.Paid != 0 || md.Owes != 3 {
			t.Error("Wrong value", diff)
		}
	}

	{  // Test decrease in Paid & Owes
		previous = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 15, Owes: 9},
		}
		current = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 4},
		}
		if diff := current.BillBalanceDifference(previous); len(diff) != 1 {
			t.Error("Should have single item", diff)
		} else if md, ok := diff["m1"]; !ok {
			t.Error("Item should be m1", diff)
		} else if md.Paid != -3 || md.Owes != -5 {
			t.Error("Wrong value", diff)
		}
	}

	{  // Test in member added
		previous = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 12},
		}
		current = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 6},
			"m2": BillMemberBalance{Paid: 0, Owes: 6},
		}
		if diff := current.BillBalanceDifference(previous); len(diff) != 2 {
			t.Error("Should have 2 items", diff)
		} else if m1, ok := diff["m1"]; !ok {
			t.Error("Item should be m1", diff)
		} else if m2, ok := diff["m2"]; !ok {
			t.Error("Item should be m1", diff)
		} else if m1.Paid != 0 || m1.Owes != -6 {
			t.Error("Wrong m1 diff", m1)
		} else if m2.Paid != 0 || m2.Owes != 6 {
			t.Error("Wrong m2 diff", m2)
		}
	}

	{  // Test in member changed
		previous = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 6},
			"m2": BillMemberBalance{Paid: 0, Owes: 6},
		}
		current = BillBalanceByMember{
			"m1": BillMemberBalance{Paid: 12, Owes: 6},
			"m3": BillMemberBalance{Paid: 0, Owes: 6},
		}
		if diff := current.BillBalanceDifference(previous); len(diff) != 2 {
			t.Error("Should have 2 items", diff)
		} else if m2, ok := diff["m2"]; !ok {
			t.Error("Item should be m1", diff)
		} else if m3, ok := diff["m3"]; !ok {
			t.Error("Item should be m1", diff)
		} else if m2.Paid != 0 || m2.Owes != -6 {
			t.Error("Wrong m2 diff", m2)
		} else if m3.Paid != 0 || m3.Owes != 6 {
			t.Error("Wrong m3 diff", m3)
		}
	}
}


func TestBillBalanceDifference_IsAffectingGroupBalance(t *testing.T) {
	var diff BillBalanceDifference

	{	// verify empty
		diff = BillBalanceDifference{}
		if diff.IsAffectingGroupBalance() {
			t.Errorf("should be false for empty map")
		}
	}

	{	// verify paid=owes for single member
		diff = BillBalanceDifference{
			"m1": BillMemberBalance{Paid: 10, Owes: 10},
		}
		if diff.IsAffectingGroupBalance() {
			t.Errorf("should be false for empty map")
		}
	}

	{	// verify paid=owes for 2 members
		diff = BillBalanceDifference{
			"m1": BillMemberBalance{Paid: 10, Owes: 10},
			"m2": BillMemberBalance{Paid: 5, Owes: 5},
		}
		if diff.IsAffectingGroupBalance() {
			t.Errorf("should be false for empty map")
		}
	}
}