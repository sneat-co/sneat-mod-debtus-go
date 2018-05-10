package facade

import (
	"testing"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dtmocks"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	is2 "github.com/matryer/is"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"strings"
	"github.com/crediterra/go-interest"
)

type assertHelper struct {
	t *testing.T
}

func (assert assertHelper) OutputIsNilIfErr(output createTransferOutput, err error) (createTransferOutput, error) {
	assert.t.Helper()
	if err != nil {
		if output.Transfer.ID != 0 {
			assert.t.Errorf("Returned transfer.ID != 0 with error: (ID=%v), error: %v", output.Transfer.ID, err)
		}
		if output.Transfer.TransferEntity != nil {
			assert.t.Errorf("Returned a non nil transfer entity with error: %v", err)
		}
		// if counterparty != nil {
		// 	t.Errorf("Returned a counterparty with error: %v", counterparty)
		// }
		// if creatorUser != nil {
		// 	t.Errorf("Returned creatorUser with error: %v", creatorUser)
		// 	return
		// }
	}
	return output, err
}

func TestCreateTransfer(t *testing.T) {
	// c, done, err := aetest.NewContext()
	// if err != nil {
	// 	t.Fatal(err.Error())
	// }
	// defer done()
	dtmocks.SetupMocks(context.Background())
	c := context.Background()
	assert := assertHelper{t: t}

	source := dal.NewTransferSourceBot(telegram.PlatformID, "test-bot", "444")

	currency := models.CURRENCY_EUR

	const (
		userID         = 1
		counterpartyID = 2
	)

	/* Test CreateTransfer that should succeed  - new counterparty by name */
	{
		// counterparty, err := dal.Contact.CreateContact(c, user.ID, 0, 0, models.ContactDetails{
		// 	FirstName: "First",
		// 	LastName:  "Contact",
		// })

		from := &models.TransferCounterpartyInfo{
			UserID:  userID,
			Note:    "Note 1",
			Comment: "Comment 1",
		}

		to := &models.TransferCounterpartyInfo{
			ContactID: counterpartyID,
		}

		creatorUser := models.AppUser{
			IntegerID:     db.IntegerID{ID: userID},
			AppUserEntity: &models.AppUserEntity{},
		}
		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			false,
			0,
			from, to,
			models.NewAmount(currency, 10),
			time.Now().Add(time.Minute), models.NoInterest())

		output, err := assert.OutputIsNilIfErr(Transfers.CreateTransfer(c, newTransfer))
		if err != nil {
			t.Error("Should not fail:", err)
			return
		}
		transfer := output.Transfer
		fromUser, toUser := output.From.User, output.To.User
		fromCounterparty, toCounterparty := output.From.Contact, output.To.Contact

		if output.Transfer.ID == 0 {
			t.Error("transfer.ID == 0")
			return
		}
		if output.Transfer.TransferEntity == nil {
			t.Error("transfer.TransferEntity == nil")
			return
		}

		if toCounterparty.ContactEntity == nil {
			t.Error("toCounterparty.ContactEntity == nil")
			return
		}
		if fromUser.AppUserEntity == nil {
			t.Error("fromUser.AppUserEntity == nil")
			return
		}
		if transfer.CreatorUserID != userID {
			t.Errorf("transfer.CreatorUserID:%v != userID:%v", transfer.CreatorUserID, userID)
		}
		if transfer.Counterparty().ContactID != to.ContactID {
			t.Errorf("transfer.Contact().ContactID:%v != to.ContactID:%v", transfer.Counterparty().ContactID, to.ContactID)
		}
		if transfer.Counterparty().ContactName == "" {
			t.Error("transfer.Contact().ContactName is empty string")
		}

		transfer2, err := GetTransferByID(c, transfer.ID)
		if err != nil {
			t.Error(errors.Wrapf(err, "Failed to get transfer by id=%v", transfer.ID))
			return
		}
		if transfer.TransferEntity == nil {
			t.Error("transfer == nil")
			return
		}
		if len(transfer2.BothUserIDs) != 2 {
			t.Errorf("len(transfer2.BothUserIDs):%v != 2", len(transfer2.BothUserIDs))
			return
		}
		if fromUser.ID != 0 && fromUser.AppUserEntity == nil {
			t.Error("fromUser.ID != 0 && fromUser.AppUserEntity == nil")
		}
		if toUser.ID != 0 && toUser.AppUserEntity == nil {
			t.Error("toUser.ID != 0 && toUser.AppUserEntity == nil")
		}
		if toCounterparty.ID != 0 && toCounterparty.ContactEntity == nil {
			t.Error("fromCounterparty.ContactEntity == nil")
		}
		if fromCounterparty.ID != 0 && fromCounterparty.ContactEntity == nil {
			t.Error("fromCounterparty.ID != 0 && fromCounterparty.ContactEntity == nil")
		}
	}
}

func TestCreateReturnTransferWithInterest(t *testing.T) {
	c := context.TODO()
	dtmocks.SetupMocks(c)
	assert := assertHelper{t: t}
	currency := models.CURRENCY_EUR

	const (
		userID    int64 = 1
		contactID int64 = 2
	)

	var (
		output     createTransferOutput
		err        error
	)

	creatorUser, err := User.GetUserByID(c, userID)
	if err != nil {
		t.Fatal(err)
	}

	source := dal.NewTransferSourceBot(telegram.PlatformID, "test-bot", "444")

	t1val := decimal.FromInt(10)
	{ // Create 1st "gave" transfer
		from := &models.TransferCounterpartyInfo{
			UserID: userID,
		}

		to := &models.TransferCounterpartyInfo{
			ContactID: contactID,
		}

		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			false,
			0,
			from, to,
			models.NewAmount(currency, t1val),
			time.Now().Add(time.Minute), models.NoInterest())

		output, err = Transfers.CreateTransfer(c, newTransfer)

		if output, err = assert.OutputIsNilIfErr(output, err); err != nil {
			t.Fatal(err)
		}
	}

	{ // Create return transfer with interest
		from := &models.TransferCounterpartyInfo{
			ContactID: contactID,
		}

		to := &models.TransferCounterpartyInfo{
			UserID: userID,
		}

		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			false,
			0,
			from, to,
			models.NewAmount(currency, 1000),
			time.Now().Add(time.Minute), models.NewInterest(interest.FormulaSimple, 2.00, interest.RatePeriodDaily))

		output, err = Transfers.CreateTransfer(c, newTransfer)

		if err == nil {
			t.Fatal("") // should fail
		}

		if !strings.Contains(err.Error(), "interest") {
			t.Fatalf("error should mention 'interest', got: %v", err)
		}
		if !strings.Contains(err.Error(), "outstanding") {
			t.Fatalf("error should mention 'outstanding', got: %v", err)
		}
	}
}

func TestCreateTransfer_GaveGotAndFullReturn(t *testing.T) {
	c := context.TODO()
	dtmocks.SetupMocks(c)
	assert := assertHelper{t: t}
	is := is2.New(t)
	currency := models.CURRENCY_EUR

	const (
		zero      int64 = 0
		userID    int64 = 1
		contactID int64 = 2
	)

	var (
		output     createTransferOutput
		t1, t2, t3 models.Transfer
		err        error
	)

	creatorUser, err := User.GetUserByID(c, userID)
	if err != nil {
		t.Fatal(err)
	}

	source := dal.NewTransferSourceBot(telegram.PlatformID, "test-bot", "444")

	t1val := decimal.FromInt(10)
	{ // Create 1st "gave" transfer
		from := &models.TransferCounterpartyInfo{
			UserID: userID,
		}

		to := &models.TransferCounterpartyInfo{
			ContactID: contactID,
		}

		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			false,
			0,
			from, to,
			models.NewAmount(currency, t1val),
			time.Now().Add(time.Minute), models.NoInterest())

		output, err = Transfers.CreateTransfer(c, newTransfer)

		if output, err = assert.OutputIsNilIfErr(output, err); err != nil {
			t.Fatal(err)
		}
		t1 = output.Transfer
		is.True(t1.ID != 0)
		is.True(t1.IsOutstanding)
		is.Equal(t1.AmountReturned, decimal.Decimal64p2(0))
		is.Equal(t1.AmountInCents, t1val)
		is.Equal(t1.GetOutstandingValue(time.Now()), t1val)
		is.Equal(output.From.User.ID, userID)
		is.Equal(output.To.Contact.ID, contactID)
		is.Equal(output.From.Contact.ID, zero)
		is.Equal(output.To.User.ID, zero)
		is.Equal(output.To.Contact.BalanceCount, 1)
		is.Equal(len(output.To.Contact.Balance()), 1)
		is.Equal(output.From.User.BalanceCount, 1)
		is.Equal(len(output.From.User.Balance()), 1)
		is.Equal(output.From.User.Balance()[currency], t1val)
		is.Equal(output.To.Contact.Balance()[currency], t1val)

		is.Equal(output.From.User.ContactsCountActive, 1)
		is.Equal(len(output.From.User.ActiveContacts()), 1)
		is.Equal(len(output.From.User.ActiveContacts()[0].Balance()), 1)
	}

	creatorUser = output.From.User

	t2val := decimal.NewDecimal64p2FromFloat64(17.00)
	{ // Create 2nd got transfer
		from := &models.TransferCounterpartyInfo{
			ContactID: contactID,
		}

		to := &models.TransferCounterpartyInfo{
			UserID: userID,
		}

		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			false,
			0,
			from, to,
			models.NewAmount(currency, t2val),
			time.Now().Add(time.Minute), models.NoInterest())

		output, err = Transfers.CreateTransfer(c, newTransfer)

		if output, err = assert.OutputIsNilIfErr(output, err); err != nil {
			t.Fatal(err)
			return
		}

		t2 = output.Transfer
		is.True(t2.ID != zero)
		is.Equal(output.To.Contact.ID, zero)
		is.Equal(output.From.User.ID, zero)
		is.True(!t1.IsOutstanding) // 1st transfer should be closed
		is.True(t2.IsOutstanding)  // 2nd transfer should be outstanding
		is.Equal(t2.AmountInCents, t2val)
		is.Equal(t2.AmountReturned, t1val)
		is.Equal(t1.AmountReturned, t1val)

		{
			toUser := output.To.User
			is.Equal(toUser.ID, userID)
			is.Equal(toUser.BalanceCount, 1)
			is.Equal(len(toUser.Balance()), 1)
			is.Equal(toUser.Balance()[currency], decimal.NewDecimal64p2FromFloat64(-7.00))
		}
		{
			fromContact := output.From.Contact
			is.Equal(fromContact.ID, contactID)
			is.Equal(fromContact.BalanceCount, 1)
			is.Equal(len(fromContact.Balance()), 1)
			is.Equal(fromContact.Balance()[currency], decimal.NewDecimal64p2FromFloat64(-7.00))
		}
	}

	if t1, err = GetTransferByID(c, t1.ID); err != nil {
		t.Fatal(err)
	}
	if t2, err = GetTransferByID(c, t2.ID); err != nil {
		t.Fatal(err)
	}

	t3val := decimal.NewDecimal64p2FromFloat64(7.00)
	{ // Create 3d transfer - full return
		from := &models.TransferCounterpartyInfo{
			UserID: userID,
		}

		to := &models.TransferCounterpartyInfo{
			ContactID: contactID,
		}

		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			true,
			0,
			from, to,
			models.NewAmount(currency, t3val),
			time.Now().Add(time.Minute), models.NoInterest())

		output, err = Transfers.CreateTransfer(c, newTransfer)
		if output, err = assert.OutputIsNilIfErr(output, err); err != nil {
			t.Errorf(err.Error())
			return
		}
		t3 = output.Transfer
		is.True(t3.ID != zero)
		is.Equal(output.From.User.ID, userID)
		is.Equal(output.To.Contact.ID, contactID)
		is.Equal(output.From.Contact.ID, zero)
		is.Equal(output.To.User.ID, zero)

		{
			fromUser := output.From.User
			is.Equal(fromUser.BalanceCount, 0)
			is.Equal(len(fromUser.Balance()), 0)
		}
		{
			toContact := output.To.Contact
			is.Equal(toContact.BalanceCount, 0)
			is.Equal(len(toContact.Balance()), 0)
		}

		is.True(!t1.IsOutstanding)
		is.True(!t2.IsOutstanding)
		is.True(!t3.IsOutstanding)

		is.Equal(t2.AmountReturned, t2val) // t2.AmountReturned
		is.Equal(t2.GetOutstandingValue(time.Now()), decimal.FromInt(0))

		is.Equal(t3.AmountReturned, decimal.FromInt(7)) // t3.AmountReturned
		is.Equal(t3.GetOutstandingValue(time.Now()), decimal.FromInt(0))
	}

	// println("t1", t1.String())
	// println("t2", t2.String())
	// println("t3", t3.String())
}

func Test_removeClosedTransfersFromOutstandingWithInterest(t *testing.T) {
	transfersWithInterest := []models.TransferWithInterestJson{
		{TransferID: 1},
		{TransferID: 2},
		{TransferID: 3},
		{TransferID: 4},
		{TransferID: 5},
	}
	transfersWithInterest = removeClosedTransfersFromOutstandingWithInterest(transfersWithInterest, []int64{2, 3})
	if len(transfersWithInterest) != 3 {
		t.Fatalf("len(transfersWithInterest) != 3: %v", transfersWithInterest)
	}
	for i, transferID := range []int64{1, 4, 5} {
		if transfersWithInterest[i].TransferID != transferID {
			t.Fatalf("transfersWithInterest[%v].TransferID: %v != %v", i, transfersWithInterest[i].TransferID, transferID)
		}
	}
}
