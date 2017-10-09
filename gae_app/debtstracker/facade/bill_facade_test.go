package facade

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal/dalmocks"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/decimal"
	"golang.org/x/net/context"
	"strings"
	"testing"
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
	Bill.CreateBill(context.Background(), context.Background(), nil)
}

func TestCreateBillErrorNoMembers(t *testing.T) {
	dal.Bill = dalmocks.NewBillDalMock()
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.SetBillMembers([]models.BillMemberJson{})
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err != nil {
		if !strings.Contains(err.Error(), "members") {
			t.Error("Error does not mention members:", err)
			return
		}
	}
	if bill.ID != 1 {
		t.Error("Unexpected bill ID:", bill.ID)
		return
	}

}

const mockBillID = 1

func TestCreateBillAmountZeroError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.AmountTotal = 0
	billEntity.Currency = "EUR"
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "== 0") || !strings.Contains(errText, "AmountTotal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillAmountNegativeError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.AmountTotal = -5
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "< 0") || !strings.Contains(errText, "AmountTotal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillAmountError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	billEntity.AmountTotal += 5
	members[0].Paid += 5
	billEntity.SetBillMembers(members)
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "totalOwedByMembers != billEntity.AmountTotal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillStatusMissingError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.Status = ""
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "required") || !strings.Contains(errText, "Status") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillStatusUnknownError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	billEntity.Status = "bogus"
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "Invalid status") || !strings.Contains(errText, "expected one of") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillMemberNegativeAmountError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	members[3].Owes *= -1
	billEntity.SetBillMembers(members)
	billEntity.AmountTotal += members[3].Owes
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "negative") || !strings.Contains(errText, "members[3]") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillTooManyMembersError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	members[0].Paid = billEntity.AmountTotal / 2
	members[1].Paid = billEntity.AmountTotal / 2
	members[2].Paid = billEntity.AmountTotal / 2
	if err := billEntity.SetBillMembers(members); err != nil {
		t.Error(err)
	}
	c := context.Background()
	bill, err := Bill.CreateBill(c, c, billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if errText != "bill has too many payers" {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillMembersOverPaid(t *testing.T) {
	SetupMocks(context.Background())
	billEntity := createGoodBillSplitByPercentage(t)
	members := billEntity.GetBillMembers()
	members[0].Paid = billEntity.AmountTotal + 10
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
	}
	errText := err.Error()
	if !strings.Contains(errText, "Total paid") || !strings.Contains(errText, "equal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillSuccess(t *testing.T) {
	mockDB := SetupMocks(context.Background())

	billEntity := createGoodBillSplitByPercentage(t)

	members := billEntity.GetBillMembers()
	//t.Logf("billEntity.GetBillMembers(): %v", members)

	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID != mockBillID {
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

	var verifyMember = func(i int, expectedUserID int64) {
		member := members[i]
		if member.UserID != int64(expectedUserID) {
			t.Errorf("members[%d].UserID == %d, expected: %d, member: %v", i, member.UserID, expectedUserID, member)
		}
	}
	verifyMember(0, 1)
	verifyMember(1, 3)
	verifyMember(2, 5)
	verifyMember(3, 0)

	if len(mockDB.BillMock.Bills) != 1 {
		t.Errorf("Expected to have 1 bill in DB, got: %d", len(mockDB.BillMock.Bills))
	}
}

func createGoodBillSplitByPercentage(t *testing.T) (billEntity *models.BillEntity) {
	billEntity = new(models.BillEntity)
	billEntity.Status = models.BillStatusActive
	billEntity.SplitMode = models.SplitModePercentage
	billEntity.CreatorUserID = 1
	billEntity.AmountTotal = 848
	billEntity.Currency = "EUR"

	percent := 25
	if err := billEntity.SetBillMembers([]models.BillMemberJson{
		{Owes: 212, MemberJson: models.MemberJson{ID: "1", Shares: percent, UserID: 1, Name: "First member"}, Paid: billEntity.AmountTotal},
		{Owes: 212, MemberJson: models.MemberJson{ID: "2", Shares: percent, UserID: 3, Name: "Second contact", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 2, ContactName: "Second contact"}}}},
		{Owes: 212, MemberJson: models.MemberJson{ID: "3", Shares: percent, UserID: 5, Name: "Fifth user", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 4, ContactName: "Forth contact"}}}},
		{Owes: 212, MemberJson: models.MemberJson{ID: "4", Shares: percent, Name: "12th contact", ContactByUser: models.MemberContactsJsonByUser{"5": models.MemberContactJson{ContactID: 12, ContactName: "12th contact"}}}},
	}); err != nil {
		t.Error(errors.WithMessage(err, "Failed to set members"))
		return
	}
	return
}

func createGoodBillSplitEqually(t *testing.T) (billEntity *models.BillEntity, err error) {
	billEntity = new(models.BillEntity)
	billEntity.Status = models.BillStatusActive
	billEntity.SplitMode = models.SplitModeEqually
	billEntity.CreatorUserID = 1
	billEntity.AmountTotal = 637
	billEntity.Currency = "EUR"

	if err = billEntity.SetBillMembers([]models.BillMemberJson{
		{Owes: 213, MemberJson: models.MemberJson{ID: "1", UserID: 1, Name: "First user"}, Paid: billEntity.AmountTotal},
		{Owes: 212, MemberJson: models.MemberJson{ID: "2", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 2}}}},
		{Owes: 212, MemberJson: models.MemberJson{ID: "3", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 4}}}},
	}); err != nil {
		err = errors.WithMessage(err, "Failed to set members")
		return
	}
	return
}

func createGoodBillSplitByShare(t *testing.T) (billEntity *models.BillEntity, err error) {
	billEntity = new(models.BillEntity)
	billEntity.Status = models.BillStatusActive
	billEntity.SplitMode = models.SplitModeShare
	billEntity.CreatorUserID = 1
	billEntity.AmountTotal = 636
	billEntity.Currency = "EUR"

	if err = billEntity.SetBillMembers([]models.BillMemberJson{
		{Owes: 212, MemberJson: models.MemberJson{ID: "1", Shares: 2, UserID: 1, Name: "First user"}, Paid: billEntity.AmountTotal},
		{Owes: 106, MemberJson: models.MemberJson{ID: "2", Shares: 1, ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 2}}}},
		{Owes: 318, MemberJson: models.MemberJson{ID: "3", Shares: 3, ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 4}}}},
	}); err != nil {
		t.Error(errors.WithMessage(err, "Failed to set members"))
		return
	}
	return
}

func createGoodBillSplitWithAdjustments(t *testing.T) (billEntity *models.BillEntity, err error) {
	billEntity = new(models.BillEntity)
	billEntity.Status = models.BillStatusActive
	billEntity.SplitMode = models.SplitModeAdjustment
	billEntity.CreatorUserID = 1
	billEntity.AmountTotal = 636
	billEntity.Currency = "EUR"

	if err = billEntity.SetBillMembers([]models.BillMemberJson{
		{Owes: 202, MemberJson: models.MemberJson{ID: "1", UserID: 1, Name: "First user"}, Paid: billEntity.AmountTotal},
		{Owes: 212, MemberJson: models.MemberJson{ID: "2", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 2}}}, Adjustment: 10},
		{Owes: 222, MemberJson: models.MemberJson{ID: "3", ContactByUser: models.MemberContactsJsonByUser{"1": models.MemberContactJson{ContactID: 4}}}, Adjustment: 20},
	}); err != nil {
		t.Error(errors.WithMessage(err, "Failed to set members"))
		return
	}
	return
}

func TestCreateBillEquallyTooManyAmountsError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitEqually(t)
	if err != nil {
		t.Error(err)
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Owes -= decimal.NewDecimal64p2FromFloat64(0.01)
	billEntity.SetBillMembers(members)
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "len(amountsCountByValue) > 2") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillEquallyAmountDeviateTooMuchError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitEqually(t)
	if err != nil {
		t.Error(err)
		return
	}
	members := billEntity.GetBillMembers()
	members[0].Owes += decimal.NewDecimal64p2FromFloat64(0.01)
	members[1].Owes -= decimal.NewDecimal64p2FromFloat64(0.01)
	billEntity.SetBillMembers(members)
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "deviated too much") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillEquallySuccess(t *testing.T) {
	c := context.Background()
	SetupMocks(c)
	billEntity, err := createGoodBillSplitEqually(t)
	if err != nil {
		t.Error(err)
		return
	}
	bill, err := Bill.CreateBill(c, c, billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID == 0 {
		t.Error("bill.ID == 0")
	}
}

func TestCreateBillAdjustmentSuccess(t *testing.T) {
	c := context.Background()
	SetupMocks(c)
	billEntity, err := createGoodBillSplitWithAdjustments(t)
	if err != nil {
		t.Error(err)
		return
	}
	bill, err := Bill.CreateBill(c, c, billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID == 0 {
		t.Error("bill.ID == 0")
	}
}

func TestCreateBillAdjustmentTotalAdjustmentIsTooBigError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitWithAdjustments(t)
	if err != nil {
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Adjustment += decimal.NewDecimal64p2FromFloat64(4.15)
	members[2].Adjustment += decimal.NewDecimal64p2FromFloat64(3.16)
	billEntity.SetBillMembers(members)
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "totalAdjustmentByMembers > billEntity.AmountTotal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillAdjustmentMemberAdjustmentIsTooBigError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitWithAdjustments(t)
	if err != nil {
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Adjustment += decimal.NewDecimal64p2FromFloat64(7.19)
	billEntity.SetBillMembers(members)
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "members[1].Adjustment > billEntity.AmountTotal") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillAdjustmentAmountDeviateTooMuchError(t *testing.T) {
	SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitWithAdjustments(t)
	if err != nil {
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Adjustment += decimal.NewDecimal64p2FromFloat64(0.10)
	billEntity.SetBillMembers(members)
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "deviated too much") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
}

func TestCreateBillShareSuccess(t *testing.T) {
	SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitByShare(t)
	if err != nil {
		return
	}
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err != nil {
		t.Error(err)
		return
	}
	if bill.ID == 0 {
		t.Error("bill.ID == 0")
	}
}

func TestCreateBillShareAmountDeviateTooMuchError(t *testing.T) {
	mockDB := SetupMocks(context.Background())
	billEntity, err := createGoodBillSplitWithAdjustments(t)
	if err != nil {
		return
	}
	members := billEntity.GetBillMembers()
	members[1].Owes += decimal.NewDecimal64p2FromFloat64(0.10)
	members[2].Owes -= decimal.NewDecimal64p2FromFloat64(0.10)
	billEntity.SetBillMembers(members)
	bill, err := Bill.CreateBill(context.Background(), context.Background(), billEntity)
	if err == nil {
		t.Error("Error expected")
		return
	}
	errText := err.Error()
	if !strings.Contains(errText, "deviated too much") {
		t.Error("Unexpected error text:", errText)
	}
	if bill.ID != 0 {
		t.Error("bill.ID != 0")
	}
	if len(mockDB.BillMock.Bills) != 0 {
		t.Errorf("Expected to have 0 bills in database, got: %d", len(mockDB.BillMock.Bills))
	}
}
