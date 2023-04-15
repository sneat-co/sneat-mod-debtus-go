package facade

import (
	// "bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	// "bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal/dalmocks"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"

	// "errors"
	"strings"
	"testing"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtmocks"
	"context"
	"github.com/strongo/decimal"
)

func TestCreateBillPanicOnNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic")
		} else {
			err := r.(string)
			if !strings.Contains(err, "context.Context") {
				t.Errorf("Error does not mention context: %v", err)
			}
		}
	}()
	Bill.CreateBill(nil, nil, nil)
}

func TestCreateBillPanicOnNilBill(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic")
		} else {
			err := r.(string)
			if !strings.Contains(err, "*models.BillEntity") {
				t.Errorf("Error does not mention bill: %v", err)
			}
		}
	}()
	_, _ = Bill.CreateBill(context.Background(), nil, nil)
}

func TestCreateBillErrorNoMembers(t *testing.T) {
	// dtdal.Bill = dalmocks.NewBillDalMock()
	// billEntity := createGoodBillSplitByPercentage(t)
	// billEntity.setBillMembers([]models.BillMemberJson{})
	// bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	// if err != nil {
	// 	if !strings.Contains(err.Error(), "members") {
	// 		t.Error("Error does not mention members:", err)
	// 		return
	// 	}
	// }
	// if bill.ID != 1 {
	// 	t.Error("Unexpected bill ID:", bill.ID)
	// 	return
	// }

}

const mockBillID = "1"

func TestCreateBillAmountZeroError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.AmountTotal = 0
	billEntity.Currency = "EUR"
	bill, err := Bill.CreateBill(context.Background(), nil, billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "== 0") || !strings.Contains(errText, "AmountTotal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != "" {
		t.Error("bill.ID != empty string")
	}
}

func TestCreateBillAmountNegativeError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.AmountTotal = -5
	bill, err := Bill.CreateBill(context.Background(), nil, billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "< 0") || !strings.Contains(errText, "AmountTotal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != "" {
		t.Error("bill.ID != empty string")
	}
}

func TestCreateBillAmountError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	billEntity.AmountTotal += 5
	members[0].Paid += 5
	// billEntity.setBillMembers(members)
	// bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	// if err == nil {
	// 	t.Error("Error expected")
	// }
	// errText := err.Error()
	// if !strings.Contains(errText, "totalOwedByMembers != billEntity.AmountTotal") {
	// 	t.Error("Unexpected error text:", errText)
	// }
	// if bill.ID != 0 {
	// 	t.Error("bill.ID != 0")
	// }
}

func TestCreateBillStatusMissingError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.Status = ""
	bill, err := Bill.CreateBill(context.Background(), nil, billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "required") || !strings.Contains(errText, "Status") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != "" {
		t.Error("bill.ID != empty string")
	}
}

func TestCreateBillStatusUnknownError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.Status = "bogus"
	bill, err := Bill.CreateBill(context.Background(), nil, billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "Invalid status") || !strings.Contains(errText, "expected one of") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != "" {
		t.Error("bill.ID != empty string")
	}
}

func TestCreateBillMemberNegativeAmountError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	members[3].Owes *= -1
	// billEntity.setBillMembers(members)
	// billEntity.AmountTotal += members[3].Owes
	// bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	// if err == nil {
	// 	t.Error("Error expected")
	// }
	// errText := err.Error()
	// if !strings.Contains(errText, "negative") || !strings.Contains(errText, "members[3]") {
	// 	t.Error("Unexpected error text:", errText)
	// }
	// if bill.ID != 0 {
	// 	t.Error("bill.ID != 0")
	// }
}

func TestCreateBillTooManyMembersError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	members[0].Paid = billEntity.AmountTotal / 2
	members[1].Paid = billEntity.AmountTotal / 2
	members[2].Paid = billEntity.AmountTotal / 2
	if err := billEntity.SetBillMembers(members); err != nil {
		t.Error(err)
	}
	c := context.Background()
	bill, err := Bill.CreateBill(c, nil, billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if errText != "bill has too many payers" {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID == "0" {
		t.Error("bill.ID is empty string")
	}
}

func TestCreateBillMembersOverPaid(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	members[0].Paid = billEntity.AmountTotal + 10
	if err := billEntity.SetBillMembers(members); err != nil {
		t.Fatal(err)
	}
	bill, err := Bill.CreateBill(context.Background(), nil, billEntity)
	if err == nil {
		t.Fatal("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "Total paid") || !strings.Contains(errText, "equal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != "" {
		t.Error("bill.ID != empty string")
	}
}

var verifyMemberUserID = func(t *testing.T, members []models.BillMemberJson, i int, expectedUserID string) {
	member := members[i]
	if member.UserID != expectedUserID {
		t.Errorf("members[%d].UserID == %v, expected: %v, member: %+v", i, member.UserID, expectedUserID, member)
	}
}

var verifyMemberOwes = func(t *testing.T, members []models.BillMemberJson, i int, expecting decimal.Decimal64p2) {
	member := members[i]
	if member.Owes != expecting {
		t.Errorf("members[%d].Owes:%v == %v", i, member.Owes, expecting)
	}
}

func TestCreateBillSuccess(t *testing.T) {
	c := context.Background()
	dtmocks.SetupMocks(c)
	billEntity := createGoodBillSplitByPercentage(t)

	members := billEntity.GetBillMembers()

	bill, err := Bill.CreateBill(c, nil, billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID == "" {
		t.Error("Unexpected bill ID", bill.ID)
		return
	}

	members = billEntity.GetBillMembers()
	if err != nil {
		t.Error(err)
		return
	}
	if len(members) != billEntity.MembersCount {
		t.Error("len(members) != billEntity.MembersCount")
	}

	verifyMemberUserID(t, members, 0, "1")
	verifyMemberUserID(t, members, 1, "3")
	verifyMemberUserID(t, members, 2, "5")
	verifyMemberUserID(t, members, 3, "")

	// if len(mockDB.BillMock.Bills) != 1 {
	// 	t.Errorf("Expected to have 1 bill in DB, got: %d", len(mockDB.BillMock.Bills))
	// }
}

func createGoodBillSplitByPercentage(t *testing.T) (billEntity *models.BillEntity) {
	billEntity = new(models.BillEntity)
	billEntity.Status = models.BillStatusOutstanding
	billEntity.SplitMode = models.SplitModePercentage
	billEntity.CreatorUserID = "1"
	billEntity.AmountTotal = 848
	billEntity.Currency = "EUR"

	percent := 25
	if err := billEntity.SetBillMembers([]models.BillMemberJson{
		{Percent: 2500, MemberJson: models.MemberJson{ID: "1", Shares: percent, UserID: "1", Name: "First member"}, Paid: billEntity.AmountTotal},
		{Percent: 2500, MemberJson: models.MemberJson{ID: "2", Shares: percent, UserID: "3", Name: "Second contact", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: "2", ContactName: "Second contact"}}}},
		{Percent: 2500, MemberJson: models.MemberJson{ID: "3", Shares: percent, UserID: "5", Name: "Fifth user", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: "4", ContactName: "Forth contact"}}}},
		{Percent: 2500, MemberJson: models.MemberJson{ID: "4", Shares: percent, Name: "12th contact", ContactByUser: models.MemberContactsJsonByUser{"5": models.MemberContactJson{ContactID: "12", ContactName: "12th contact"}}}},
	}); err != nil {
		t.Error(fmt.Errorf("%w: Failed to set members", err))
		return
	}
	return
}

func createGoodBillSplitEqually(t *testing.T) (billEntity *models.BillEntity, err error) {
	billEntity = new(models.BillEntity)
	billEntity.Status = models.BillStatusOutstanding
	billEntity.SplitMode = models.SplitModeEqually
	billEntity.CreatorUserID = "1"
	billEntity.AmountTotal = 637
	billEntity.Currency = "EUR"

	if err = billEntity.SetBillMembers([]models.BillMemberJson{
		{Owes: 213, MemberJson: models.MemberJson{ID: "1", UserID: "1", Name: "First user"}, Paid: billEntity.AmountTotal},
		{Owes: 212, MemberJson: models.MemberJson{ID: "2", Name: "Second", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: "2"}}}},
		{Owes: 212, MemberJson: models.MemberJson{ID: "3", Name: "Forth", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: "4"}}}},
	}); err != nil {
		err = fmt.Errorf("%w: Failed to set members", err)
		return
	}
	return
}

func createGoodBillSplitEquallyWithAdjustments(t *testing.T) (billEntity *models.BillEntity, err error) {
	t.Helper()

	if billEntity, err = createGoodBillSplitEqually(t); err != nil {
		return
	}

	members := billEntity.GetBillMembers()
	members[1].Adjustment = 10
	members[2].Adjustment = 20
	if err = billEntity.SetBillMembers(members); err != nil {
		t.Fatal(err)
	}
	members = billEntity.GetBillMembers()
	if len(members) != 3 {
		t.Fatal("len(members) != 3")
	}
	/*
		637 - 30 = 607
		607 / 3 = 202
	*/
	validateOwes := func(i int, expecting decimal.Decimal64p2) {
		if members[i].Owes != expecting {
			t.Fatalf("members[%d].Owes:%v != %v", i, members[0].Owes, expecting)
		}
	}
	validateOwes(0, 203)
	validateOwes(1, 212)
	validateOwes(2, 222)
	return
}

func createGoodBillSplitByShare(t *testing.T) (billEntity *models.BillEntity, err error) {
	billEntity = new(models.BillEntity)
	billEntity.Status = models.BillStatusOutstanding
	billEntity.SplitMode = models.SplitModeShare
	billEntity.CreatorUserID = "1"
	billEntity.AmountTotal = 636
	billEntity.Currency = "EUR"

	if err = billEntity.SetBillMembers([]models.BillMemberJson{
		{MemberJson: models.MemberJson{ID: "1", Shares: 2, UserID: "1", Name: "First user"}, Paid: billEntity.AmountTotal},
		{MemberJson: models.MemberJson{ID: "2", Shares: 1, Name: "Second", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: "2"}}}},
		{MemberJson: models.MemberJson{ID: "3", Shares: 3, Name: "Forth", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: "4"}}}},
	}); err != nil {
		t.Error(fmt.Errorf("%w: Failed to set members", err))
		return
	}
	members := billEntity.GetBillMembers()
	verifyMemberOwes(t, members, 0, 212)
	verifyMemberOwes(t, members, 1, 106)
	verifyMemberOwes(t, members, 2, 318)
	return
}

// There is no way to check as we do not expose membser publicly
// func TestCreateBillEquallyTooManyAmountsError(t *testing.T) {
// 	c := context.Background()
// 	dtmocks.SetupMocks(c)
// 	billEntity, err := createGoodBillSplitEqually(t)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	members := billEntity.GetBillMembers()
// 	members[1].Owes -= decimal.NewDecimal64p2FromFloat64(0.01)
// 	t.Logf("memebers: %v", members)
// 	if err = billEntity.SetBillMembers(members); err != nil {
// 		t.Fatal(err)
// 	}
// 	bill, err := Bill.CreateBill(c, c, billEntity)
// 	if err == nil {
// 		t.Fatal("Error expected")
// 	}
// 	errText := err.Error()
// 	if !strings.Contains(errText, "len(amountsCountByValue) > 2") {
// 		t.Error("Unexpected error text:", errText)
// 	}
// 	if bill.ID == "" {
// 		t.Error("bill.ID is empty string")
// 	}
// }

// func TestCreateBillEquallyAmountDeviateTooMuchError(t *testing.T) {
// 	c := context.Background()
// 	dtmocks.SetupMocks(c)
// 	billEntity, err := createGoodBillSplitEqually(t)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	members := billEntity.GetBillMembers()
// 	members[0].Owes += decimal.NewDecimal64p2FromFloat64(0.01)
// 	members[1].Owes -= decimal.NewDecimal64p2FromFloat64(0.01)
// 	if err = billEntity.SetBillMembers(members); err != nil {
// 		t.Fatal(err)
// 	}
// 	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
// 	if err == nil {
// 		t.Error("Error expected")
// 		return
// 	}
// 	errText := err.Error()
// 	if !strings.Contains(errText, "deviated too much") {
// 		t.Error("Unexpected error text:", errText)
// 	}
// 	if bill.ID == "" {
// 		t.Error("bill.ID is empty string")
// 	}
// }

func TestCreateBillEquallySuccess(t *testing.T) {
	c := context.Background()
	dtmocks.SetupMocks(c)
	billEntity, err := createGoodBillSplitEqually(t)
	if err != nil {
		t.Error(err)
		return
	}
	bill, err := Bill.CreateBill(c, nil, billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID == "" {
		t.Error("bill.ID is empty string")
	}
}

func TestCreateBillAdjustmentSuccess(t *testing.T) {
	c := context.Background()
	dtmocks.SetupMocks(c)
	billEntity, err := createGoodBillSplitEquallyWithAdjustments(t)
	if err != nil {
		t.Fatal(err)
	}
	bill, err := Bill.CreateBill(c, nil, billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID == "" {
		t.Error("bill.ID is empty string")
	}
}

func TestCreateBillAdjustmentTotalAdjustmentIsTooBigError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitEquallyWithAdjustments(t)
	if err != nil {
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Adjustment += decimal.NewDecimal64p2FromFloat64(4.15)
	members[2].Adjustment += decimal.NewDecimal64p2FromFloat64(3.16)
	// billEntity.setBillMembers(members)
	// bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	// if err == nil {
	// 	t.Error("Error expected")
	// 	return
	// }
	// errText := err.Error()
	// if !strings.Contains(errText, "totalAdjustmentByMembers > billEntity.AmountTotal") {
	// 	t.Error("Unexpected error text:", errText)
	// }
	// if bill.ID != 0 {
	// 	t.Error("bill.ID != 0")
	// }
}

func TestCreateBillAdjustmentMemberAdjustmentIsTooBigError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitEquallyWithAdjustments(t)
	if err != nil {
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Adjustment += decimal.NewDecimal64p2FromFloat64(7.19)
	// billEntity.setBillMembers(members)
	// bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	// if err == nil {
	// 	t.Error("Error expected")
	// 	return
	// }
	// errText := err.Error()
	// if !strings.Contains(errText, "members[1].Adjustment > billEntity.AmountTotal") {
	// 	t.Error("Unexpected error text:", errText)
	// }
	// if bill.ID != 0 {
	// 	t.Error("bill.ID != 0")
	// }
}

func TestCreateBillAdjustmentAmountDeviateTooMuchError(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitEquallyWithAdjustments(t)
	if err != nil {
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Adjustment += decimal.NewDecimal64p2FromFloat64(0.10)
	// billEntity.setBillMembers(members)
	// bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	// if err == nil {
	// 	t.Error("Error expected")
	// 	return
	// }
	// errText := err.Error()
	// if !strings.Contains(errText, "deviated too much") {
	// 	t.Error("Unexpected error text:", errText)
	// }
	// if bill.ID != 0 {
	// 	t.Error("bill.ID != 0")
	// }
}

func TestCreateBillShareSuccess(t *testing.T) {
	dtmocks.SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitByShare(t)
	if err != nil {
		return
	}
	bill, err := Bill.CreateBill(context.Background(), nil, billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID == "" {
		t.Error("bill.ID is empty string")
	}
}

func TestCreateBillShareAmountDeviateTooMuchError(t *testing.T) {
	// mockDB := dtmocks.SetupMocks(context.Background())
	// billEntity, err := createGoodBillSplitEquallyWithAdjustments(t)
	// if err != nil {
	// 	return
	// }
	// members := billEntity.GetBillMembers()
	// members[1].Owes += decimal.NewDecimal64p2FromFloat64(0.10)
	// members[2].Owes -= decimal.NewDecimal64p2FromFloat64(0.10)
	// billEntity.setBillMembers(members)
	// bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	// if err == nil {
	// 	t.Error("Error expected")
	// 	return
	// }
	// errText := err.Error()
	// if !strings.Contains(errText, "deviated too much") {
	// 	t.Error("Unexpected error text:", errText)
	// }
	// if bill.ID != 0 {
	// 	t.Error("bill.ID != 0")
	// }
	// if len(mockDB.BillMock.Bills) != 0 {
	// 	t.Errorf("Expected to have 0 bills in database, got: %d", len(mockDB.BillMock.Bills))
	// }
}
