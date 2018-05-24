package facade

import (
	"testing"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtmocks"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/crediterra/go-interest"
	is2 "github.com/matryer/is"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"strings"
	"runtime/debug"
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
		is.Equal(t1.AmountInCentsReturned, decimal.Decimal64p2(0))
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
		is.Equal(t2.AmountInCentsReturned, t1val)
		is.Equal(t1.AmountInCentsReturned, t1val)

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

		is.Equal(t2.AmountInCentsReturned, t2val) // t2.AmountInCentsReturned
		is.Equal(t2.GetOutstandingValue(time.Now()), decimal.FromInt(0))

		is.Equal(t3.AmountInCentsReturned, decimal.FromInt(7)) // t3.AmountInCentsReturned
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

type createTransferTestCase struct {
	name  string
	steps []createTransferStep
}

type createTransferStepInput struct {
	direction          models.TransferDirection
	isReturn           bool
	returnToTransferID int64
	amount             decimal.Decimal64p2
	time               time.Time
	models.TransferInterest
}

type createTransferStepExpects struct {
	isExpectedError func(error) bool
	balance         decimal.Decimal64p2
	amountReturned  decimal.Decimal64p2
	returns         models.TransferReturns
	transfers       []transferExpectedState
}

type createTransferStep struct {
	input             createTransferStepInput
	createdTransferID int64 // Save transfer ID for checking transfer state in later steps
	expects           createTransferStepExpects
}

type transferExpectedState struct {
	stepIndex      int
	isOutstanding  bool
	AmountReturned decimal.Decimal64p2
}

func TestCreateTransfers(t *testing.T) {
	dayHour := func(d, h int) time.Time {
		return time.Date(2000, 01, d, h, 0, 0, 0, time.Local)
	}

	expects := func(
		balance decimal.Decimal64p2,
		amountReturned decimal.Decimal64p2,
		returns models.TransferReturns,
		transfers []transferExpectedState,
	) createTransferStepExpects {
		return createTransferStepExpects{
			balance:        balance,
			amountReturned: amountReturned,
			returns:        returns,
			transfers:      transfers,
		}
	}

	expectsError := func(f func(error) bool) createTransferStepExpects {
		return createTransferStepExpects{
			isExpectedError: f,
		}
	}

	testCases := []createTransferTestCase{
		{
			name: "error on attempt to make a return with interest",
			steps: []createTransferStep{
				{
					input: createTransferStepInput{
						direction: models.TransferDirectionUser2Counterparty,
						amount:    10,
						time:      dayHour(1, 1),
					},
					expects: expects(10, 0, nil, nil),
				},
				{
					input: createTransferStepInput{
						direction:        models.TransferDirectionCounterparty2User,
						amount:           10,
						time:             dayHour(1, 1),
						TransferInterest: models.NewInterest(interest.FormulaSimple, 2.00, interest.RatePeriodDaily),
					},
					expects: expectsError(func(err error) bool {
						if err == nil {
							return false
						}
						errMsg := err.Error()

						return strings.Contains(errMsg, "interest") && strings.Contains(errMsg, "outstanding")
					}),
				},
			},
		},
		{
			name: "same_day_no_interests_2_give_1_over-return",
			steps: []createTransferStep{
				{
					input: createTransferStepInput{
						direction: models.TransferDirectionUser2Counterparty,
						amount:    10,
						time:      dayHour(1, 1)},
					expects: expects(10, 0, nil, nil),
				},
				{
					input: createTransferStepInput{
						direction: models.TransferDirectionUser2Counterparty,
						amount:    15,
						time:      dayHour(1, 2)},
					expects: expects(25, 0, nil,
						[]transferExpectedState{
							{isOutstanding: true, AmountReturned: 0},
						}),
				},
				{
					input: createTransferStepInput{
						isReturn:  false,
						direction: models.TransferDirectionCounterparty2User,
						amount:    30,
						time:      dayHour(1, 3)},
					expects: expects(-5, 25,
						models.TransferReturns{
							{Amount: 10},
							{Amount: 15},
						},
						[]transferExpectedState{
							{stepIndex: 0, AmountReturned: 10, isOutstanding: false},
							{stepIndex: 1, AmountReturned: 15, isOutstanding: false},
						}),
				},
			},
		},
	}

	for i, testCase := range testCases {
		if testCase.name == "" {
			t.Fatalf("Test case #%v has no name", i+1)
		}
		if len(testCase.steps) == 0 {
			t.Fatalf("Test case #%v has no steps", i+1)
		}
		t.Run(testCase.name, func(t *testing.T) {
			testCreateTransfer(t, testCase)
		})
	}
}

func testCreateTransfer(t *testing.T, testCase createTransferTestCase) {
	c := context.TODO()
	dtmocks.SetupMocks(c)
	assert := assertHelper{t: t}
	currency := models.CURRENCY_EUR

	source := dal.NewTransferSourceBot(telegram.PlatformID, "test-bot", "444")

	const (
		userID    = 1
		contactID = 2
	)

	creatorUser := models.AppUser{
		IntegerID:     db.IntegerID{ID: userID},
		AppUserEntity: &models.AppUserEntity{},
	}

	tUser := &models.TransferCounterpartyInfo{
		UserID: userID,
	}

	tContact := &models.TransferCounterpartyInfo{
		ContactID: contactID,
	}

	var (
		i    int
		step createTransferStep
	)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Paniced at step #%v: %v\nStack: %v", i+1, r, string(debug.Stack()))
		}
	}()

	for i, step = range testCase.steps {
		var from, to *models.TransferCounterpartyInfo
		switch step.input.direction {
		case models.TransferDirectionUser2Counterparty:
			from = tUser
			to = tContact
		case models.TransferDirectionCounterparty2User:
			from = tContact
			to = tUser
		}
		newTransfer := NewTransferInput(strongo.EnvLocal,
			source,
			creatorUser,
			"",
			step.input.isReturn,
			step.input.returnToTransferID,
			from, to,
			models.NewAmount(currency, step.input.amount),
			step.input.time, step.input.TransferInterest)

		// =============================================================
		output, err := Transfers.CreateTransfer(c, newTransfer)
		// =============================================================
		if output, err = assert.OutputIsNilIfErr(output, err); err != nil {
			if step.expects.isExpectedError(err) {
				//t.Logf("step #%v: got expected error: %v", i+1, err)
				continue
			}
			t.Errorf(err.Error())
			return
		}

		// Save transfer ID for checking transfer state in later steps
		testCase.steps[i].createdTransferID = output.Transfer.ID

		var (
			contact models.Contact
		)
		switch step.input.direction {
		case models.TransferDirectionUser2Counterparty:
			creatorUser = output.From.User
			contact = output.To.Contact
		case models.TransferDirectionCounterparty2User:
			creatorUser = output.To.User
			contact = output.From.Contact
		}
		breakSteps := false
		if balance := creatorUser.Balance()[currency]; balance != step.expects.balance {
			t.Errorf("step #%v: Expected user balance does not match actual: expected:%v != got:%v",
				i+1, step.expects.balance, balance)
			breakSteps = true
		}
		userContact := creatorUser.ContactsByID()[contact.ID]
		if balance := userContact.Balance()[currency]; balance != step.expects.balance {
			t.Errorf("step #%v: Expected userContact balance does not match actual: expected:%v != got:%v",
				i+1, step.expects.balance, balance)
			breakSteps = true
		}
		if balance := contact.Balance()[currency]; balance != step.expects.balance {
			t.Errorf("step #%v: Expected contact balance does not match actual: expected:%v != got:%v",
				i+1, step.expects.balance, balance)
			breakSteps = true
		}
		{ // verify balance counts
			var expectedBalanceCount int
			if step.expects.balance != 0 {
				expectedBalanceCount = 1
			}

			if balanceCount := len(userContact.Balance()); balanceCount != expectedBalanceCount {
				t.Errorf("step #%v: Expected userContact len(balance) does not match actual: expected:%v != got:%v",
					i+1, expectedBalanceCount, balanceCount)
				breakSteps = true
			}
			if balanceCount := len(creatorUser.Balance()); balanceCount != expectedBalanceCount {
				t.Errorf("step #%v: Expected creatorUser len(balance) does not match actual: expected:%v != got:%v",
					i+1, expectedBalanceCount, balanceCount)
				breakSteps = true
			}
		}

		var dbContact models.Contact
		if dbContact, err = GetContactByID(c, contact.ID); err != nil {
			t.Errorf("step #%v: %v", i+1, err)
			break
		}
		if balance := dbContact.Balance()[currency]; balance != step.expects.balance {
			t.Errorf("step #%v: Expected contact balance does not match actual dbContact: expected:%v != got:%v",
				i+1, step.expects.balance, balance)
			breakSteps = true
		}
		if output.Transfer.AmountInCentsReturned != step.expects.amountReturned {
			t.Errorf("step #%v: Expected transfer.AmountInCentsReturned does not match actual: expected:%v != got:%v",
				i+1, step.expects.amountReturned, output.Transfer.AmountInCentsReturned)
			breakSteps = true
		}
		if output.Transfer.ReturnsCount != len(step.expects.returns) {
			t.Errorf("step #%v: Expected transfer.ReturnsCount does not match actual: expected:%v != got:%v",
				i+1, len(step.expects.transfers), output.Transfer.ReturnsCount)
			breakSteps = true
		}
		{ // Verify returns and previous transfers state
			returns := output.Transfer.GetReturns()
			if len(returns) != output.Transfer.ReturnsCount {
				t.Errorf("step #%v: transfer.ReturnsCount is not equal to len(transfer.GetReturns()): %v != %v",
					i+1, output.Transfer.ReturnsCount, len(returns))
				continue
			}
			for j, r := range returns {
				isPreviousTransfer := false
				for k := 0; k < i; k++ {
					if r.TransferID == testCase.steps[k].createdTransferID {
						isPreviousTransfer = true
						break
					}
				}
				if !isPreviousTransfer {
					t.Errorf("step #%v: transfer.Returns[%v] references unknown transfer with ID=%v",
						i+1, j, r.TransferID)
					breakSteps = true
				}
				if r.Amount != step.expects.returns[j].Amount {
					t.Errorf("step #%v: Expected transfer.Returns[%v] does not match actual: expected:%v != got:%v",
						i+1, j, step.expects.returns[j].Amount, r.Amount)
					breakSteps = true
					break
				}
			}
			for j, expectedTransfer := range step.expects.transfers {
				var previousTransfer models.Transfer
				if previousTransfer, err = GetTransferByID(c, testCase.steps[j].createdTransferID); err != nil {
					t.Fatalf("step #%v: %v",
						i+1, err)
					breakSteps = true
					break
				}
				if previousTransfer.IsOutstanding != expectedTransfer.isOutstanding {
					t.Errorf("step #%v: previous transfer from step #%v is expected to have IsOutstanding=%v but got %v",
						i+1, j+1, expectedTransfer.isOutstanding, previousTransfer.IsOutstanding)
					breakSteps = true
					break
				}
				if previousTransfer.AmountInCentsReturned != expectedTransfer.AmountReturned {
					t.Errorf("step #%v: previous transfer from step #%v is expected to have AmountInCentsReturned=%v but got %v",
						i+1, j+1, expectedTransfer.AmountReturned, previousTransfer.AmountInCentsReturned)
					breakSteps = true
					break
				}
			}
		}
		if breakSteps {
			break
		}
	}
}
