package facade

import (
	"testing"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	is2 "github.com/matryer/is"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/decimal"
	"context"
)

type assertHelper struct {
	t *testing.T
}

func (assert assertHelper) OutputIsNilIfErr(output createTransferOutput, err error) (createTransferOutput, error) {
	if err != nil {
		if output.Transfer.ID != 0 {
			assert.t.Errorf("Returned transfer.ID != 0 with error: %v", output.Transfer.ID)
		}
		if output.Transfer.TransferEntity != nil {
			assert.t.Error("Returned a non nil transfer entity with error")
		}
		//if counterparty != nil {
		//	t.Errorf("Returned a counterparty with error: %v", counterparty)
		//}
		//if creatorUser != nil {
		//	t.Errorf("Returned creatorUser with error: %v", creatorUser)
		//	return
		//}
	}
	return output, err
}

func TestCreateTransfer(t *testing.T) {
	//c, done, err := aetest.NewContext()
	//if err != nil {
	//	t.Fatal(err.Error())
	//}
	//defer done()
	SetupMocks(context.Background())
	c := context.Background()
	assert := assertHelper{t: t}

	source := dal.NewTransferSourceBot(telegram_bot.TelegramPlatformID, "test-bot", "444")

	const (
		userID         = 1
		counterpartyID = 2
	)

	/* Test CreateTransfer that should succeed  - new counterparty by name */
	{
		//counterparty, err := dal.Contact.CreateContact(c, user.ID, 0, 0, models.ContactDetails{
		//	FirstName: "First",
		//	LastName:  "Contact",
		//})

		from := &models.TransferCounterpartyInfo{
			UserID:  userID,
			Note:    "Note 1",
			Comment: "Comment 1",
		}

		to := &models.TransferCounterpartyInfo{
			ContactID: counterpartyID,
		}

		creatorUser := models.AppUser{
			ID:            userID,
			AppUserEntity: &models.AppUserEntity{},
		}
		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			false,
			0,
			from, to,
			models.NewAmount(models.CURRENCY_RUB, 10),
			time.Now().Add(time.Minute), models.TransferInterest{})

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
		transfer2, err := dal.Transfer.GetTransferByID(c, transfer.ID)
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

func TestCreateTransfer_GaveGotAndFullReturn(t *testing.T) {
	c := context.TODO()
	SetupMocks(c)
	assert := assertHelper{t: t}
	is := is2.New(t)
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
	creatorUser := models.AppUser{
		ID:            userID,
		AppUserEntity: &models.AppUserEntity{},
	}

	source := dal.NewTransferSourceBot(telegram_bot.TelegramPlatformID, "test-bot", "444")
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
			models.NewAmount(models.CURRENCY_RUB, decimal.NewDecimal64p2FromFloat64(10.00)),
			time.Now().Add(time.Minute), models.TransferInterest{})
		//t1, _, fromUser, toUser, fromCounterparty, toCounterparty
		if output, err = assert.OutputIsNilIfErr(Transfers.CreateTransfer(c, newTransfer)); err != nil {
			t.Errorf(err.Error())
			return
		}
		t1 = output.Transfer
		is.True(t1.ID != 0)
		is.Equal(output.From.User.ID, userID)
		is.Equal(output.To.Contact.ID, contactID)
		is.Equal(output.From.Contact.ID, zero)
		is.Equal(output.To.User.ID, zero)
	}

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
			models.NewAmount(models.CURRENCY_RUB, decimal.NewDecimal64p2FromFloat64(17.00)),
			time.Now().Add(time.Minute), models.TransferInterest{})

		if output, err = assert.OutputIsNilIfErr(Transfers.CreateTransfer(c, newTransfer)); err != nil {
			t.Errorf(err.Error())
			return
		}
		t2 = output.Transfer
		is.True(t2.ID != zero)
		is.Equal(output.To.User.ID, userID)
		is.Equal(output.From.Contact.ID, contactID)
		is.Equal(output.To.Contact.ID, zero)
		is.Equal(output.From.User.ID, zero)

		balance := output.To.User.Balance()
		is.Equal(len(balance), 1)
		is.Equal(balance[models.CURRENCY_RUB], decimal.NewDecimal64p2FromFloat64(-7.00))
	}

	is.Equal(t1.AmountInCentsReturned, decimal.NewDecimal64p2FromFloat64(10))
	is.Equal(t2.AmountInCentsReturned, decimal.NewDecimal64p2FromFloat64(10))

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
			models.NewAmount(models.CURRENCY_RUB, decimal.NewDecimal64p2FromFloat64(7.00)),
			time.Now().Add(time.Minute), models.TransferInterest{})

		if output, err = assert.OutputIsNilIfErr(Transfers.CreateTransfer(c, newTransfer)); err != nil {
			t.Errorf(err.Error())
			return
		}
		t3 = output.Transfer
		is.True(t3.ID != zero)
		is.Equal(output.From.User.ID, userID)
		is.Equal(output.To.Contact.ID, contactID)
		is.Equal(output.From.Contact.ID, zero)
		is.Equal(output.To.User.ID, zero)

		balance := output.From.User.Balance()
		is.Equal(len(balance), 0)
	}

	is.Equal(t2.AmountInCentsReturned, decimal.NewDecimal64p2FromFloat64(17))
	is.Equal(t2.GetOutstandingValue(time.Now()), decimal.NewDecimal64p2FromFloat64(0))

	is.Equal(t3.AmountInCentsReturned, decimal.NewDecimal64p2FromFloat64(0))
	is.Equal(t3.GetOutstandingValue(time.Now()), decimal.NewDecimal64p2FromFloat64(0))

	println("t1", t1.String())
	println("t2", t2.String())
	println("t3", t3.String())
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
