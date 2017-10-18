package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
	"github.com/strongo/decimal"
	"golang.org/x/net/context"
	"time"
)

const (
	USER_BALANCE_INCREASED = 1
	USER_BALANCE_DECREASED = -1
)

var (
	ErrDebtAlreadyReturned                 = errors.New("This debt already has been returned")
	ErrPartialReturnGreaterThenOutstanding = errors.New("An attempt to do partial return for amount greater then outstanding")
)

func TransferCounterparties(direction models.TransferDirection, creatorInfo models.TransferCounterpartyInfo) (from, to *models.TransferCounterpartyInfo) {
	creator := models.TransferCounterpartyInfo{
		UserID:  creatorInfo.UserID,
		Comment: creatorInfo.Comment,
	}
	counterparty := models.TransferCounterpartyInfo{
		ContactID:   creatorInfo.ContactID,
		ContactName: creatorInfo.ContactName,
	}
	switch direction {
	case models.TransferDirectionUser2Counterparty:
		return &creator, &counterparty
	case models.TransferDirectionCounterparty2User:
		return &counterparty, &creator
	default:
		panic("Unknown direction: " + string(direction))
	}
}

type transferFacade struct {
}

var Transfers = transferFacade{}

type createTransferInput struct {
	Env                strongo.Environment // TODO: I believe we don't need this
	Source             dal.TransferSource
	CreatorUserID      int64
	BillID             string
	IsReturn           bool
	ReturnToTransferID int64
	//direction models.TransferDirection,
	//creatorInfo models.TransferCounterpartyInfo,
	From, To *models.TransferCounterpartyInfo
	Amount   models.Amount
	DueOn    time.Time
}

type createTransferOutputCounterparty struct {
	User    models.AppUser
	Contact models.Contact
}

type createTransferOutput struct {
	Transfer          models.Transfer
	ReturnedTransfers []models.Transfer
	From, To          *createTransferOutputCounterparty
}

func (output createTransferOutput) Validate() {
	if output.Transfer.ID == 0 {
		panic("Transfer.ID == 0")
	}
	if output.Transfer.TransferEntity == nil {
		panic("TransferEntity == nil")
	}
}

func (i createTransferInput) Validate() {
	if i.Source == nil {
		panic("source == nil")
	}
	if i.CreatorUserID == 0 {
		panic("creatorUserID == 0")
	}
	if i.Amount.Value <= 0 {
		panic("amount.Value <= 0")
	}
	if i.From == nil {
		panic("from == nil")
	}
	if i.To == nil {
		panic("to == nil")
	}

	if (i.From.ContactID == 0 && i.To.ContactID == 0) || (i.From.UserID == 0 && i.To.UserID == 0) {
		panic("(from.ContactID == 0  && to.ContactID == 0) || (from.UserID == 0 && to.UserID == 0)")
	}
	if i.From.UserID != 0 && i.To.ContactID == 0 && i.To.UserID == 0 {
		panic("from.UserID != 0 && to.ContactID == 0 && to.UserID == 0")
	}
	if i.To.UserID != 0 && i.From.ContactID == 0 && i.From.UserID == 0 {
		panic("to.UserID != 0 && from.ContactID == 0 && from.UserID == 0")
	}

	if i.From.UserID == i.To.UserID {
		if i.From.UserID == 0 && i.To.UserID == 0 {
			if i.From.ContactID == 0 {
				panic("from.UserID == 0 && to.UserID == 0 && from.ContactID == 0")
			}
			if i.To.ContactID == 0 {
				panic("from.UserID == 0 && to.UserID == 0 && to.ContactID == 0")
			}
		} else {
			panic("from.UserID == to.UserID")
		}
	}
	switch i.CreatorUserID {
	case i.From.UserID:
		if i.To.ContactID == 0 {
			panic("creatorUserID == from.UserID && to.ContactID == 0")
		}
	case i.To.UserID:
		if i.From.ContactID == 0 {
			panic("creatorUserID == from.UserID && from.ContactID == 0")
		}
	default:
		if i.From.ContactID == 0 {
			panic("3d party transfer and from.ContactID == 0")
		}
		if i.To.ContactID == 0 {
			panic("3d party transfer and to.ContactID == 0")
		}
	}
}

func (i createTransferInput) String() string {
	return fmt.Sprintf("CreatorUserID=%d, IsReturn=%v, ReturnToTransferID=%d, Amount=%v, From=%v, To=%v, DueOn=%v",
		i.CreatorUserID, i.IsReturn, i.ReturnToTransferID, i.Amount, i.From, i.To, i.DueOn)
}

func NewTransferInput(
	env strongo.Environment,
	source dal.TransferSource,
	creatorUserID int64,
	billID string,
	isReturn bool, returnToTransferID int64,
	from, to *models.TransferCounterpartyInfo,
	amount models.Amount,
	dueOn time.Time,
) (input createTransferInput) {
	input = createTransferInput{
		Env:                env,
		Source:             source,
		CreatorUserID:      creatorUserID,
		BillID:             billID,
		IsReturn:           isReturn,
		ReturnToTransferID: returnToTransferID,
		From:               from,
		To:                 to,
		Amount:             amount,
		DueOn:              dueOn,
	}
	input.Validate()
	return
}

func (transferFacade transferFacade) CreateTransfer(c context.Context, input createTransferInput) (
	output createTransferOutput, err error,
) {
	log.Infof(c, "CreateTransfer(input=%v)", input)

	originalAmountValue := input.Amount.Value
	//var counterparty *Contact
	//if creatorInfo.ContactID {
	//	counterpartyDalGae, counterparty = GetContactByID(c, creatorInfo.ContactID)
	//}

	var returnToTransferIDs []int64

	if input.ReturnToTransferID == 0 {
		if returnToTransferIDs, err = transferFacade.checkOutstandingTransfersForReturns(c, input); err != nil {
			return
		}
	} else if !input.IsReturn {
		panic("ReturnToTransferID != 0 && !IsReturn")
	} else {
		var transferToReturn models.Transfer
		if transferToReturn, err = dal.Transfer.GetTransferByID(c, input.ReturnToTransferID); err != nil {
			err = errors.Wrapf(err, "Failed to get returnToTransferID=%v", input.ReturnToTransferID)
			return
		}

		if transferToReturn.AmountInCents != transferToReturn.AmountInCentsOutstanding+transferToReturn.AmountInCentsReturned {
			panic(
				fmt.Sprintf(
					`Data integrity issue: transferToReturn.AmountTotal != transferToReturn.AmountInCentsOutstanding + transferToReturn.AmountInCentsReturned
transferToReturn.AmountTotal: %v
transferToReturn.AmountInCentsOutstanding: %v
transferToReturn.AmountInCentsReturned: %v`,
					transferToReturn.AmountInCents, transferToReturn.AmountInCentsOutstanding, transferToReturn.AmountInCentsReturned))
		}

		if transferToReturn.Currency != string(input.Amount.Currency) {
			panic("transferToReturn.Currency != amount.Currency")
		}

		if transferToReturn.AmountInCentsOutstanding == 0 {
			// When the transfer has been already returned
			err = ErrDebtAlreadyReturned
			return
		}

		if input.Amount.Value > transferToReturn.AmountInCentsOutstanding {
			log.Debugf(c, "amount.Value (%v) > transferToReturn.AmountInCentsOutstanding (%v)", input.Amount.Value, transferToReturn.AmountInCentsOutstanding)
			if input.Amount.Value == transferToReturn.AmountInCents {
				// For situations when a transfer was partially returned but user wants to mark it as fully returned.
				log.Debugf(c, "amount.Value (%v) == transferToReturn.AmountInCents (%v)", input.Amount.Value, transferToReturn.AmountInCents)
				input.Amount.Value = transferToReturn.AmountInCentsOutstanding
				log.Debugf(c, "Updated amount.Value: %v", input.Amount.Value)
			} else {
				err = ErrPartialReturnGreaterThenOutstanding
				return
			}
		} else if input.Amount.Value < transferToReturn.AmountInCentsOutstanding {
			log.Debugf(c, "amount.Value < transferToReturn.AmountInCentsOutstanding")
		}

		returnToTransferIDs = append(returnToTransferIDs, input.ReturnToTransferID)
		output.ReturnedTransfers = append(output.ReturnedTransfers, transferToReturn)
	}

	if originalAmountValue != input.Amount.Value {
		// See above for case when amount.Value changes. Logged for now to troubleshoot issue with not fully returned transfers
		log.Warningf(c, "(can be OK) originalAmountValue:%v != amount.Value:%v", originalAmountValue, input.Amount.Value)
	}

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		output, err = Transfers.createTransferWithinTransaction(c, input, returnToTransferIDs)
		return err
	}, dal.CrossGroupTransaction); err != nil {
		return
	}

	output.Validate()
	return
}

func (transferFacade transferFacade) checkOutstandingTransfersForReturns(c context.Context, input createTransferInput) (returnToTransferIDs []int64, err error) {
	var (
		outstandingTransfers []models.Transfer
	)
	outstandingTransfers, err = dal.Transfer.LoadOutstandingTransfers(c, input.CreatorUserID, input.Amount.Currency)
	if err != nil {
		err = errors.Wrap(err, "Failed to load outstanding transfers")
		return
	}
	{ // Assign the return to specific transfers
		var (
			assignedValue             decimal.Decimal64p2
			outstandingRightDirection int
		)
		var direction models.TransferDirection
		switch input.CreatorUserID {
		case input.From.UserID:
			direction = models.TransferDirectionUser2Counterparty
		case input.To.UserID:
			direction = models.TransferDirectionCounterparty2User
		default:
			if input.BillID == "" {
				panic("Not able to detect direction")
			}
		}
		for i, outstandingTransfer := range outstandingTransfers {
			if outstandingTransfer.ReturnDirectionForUser(input.CreatorUserID) == direction {
				outstandingTransferID := outstandingTransfers[i].ID
				if outstandingTransfer.AmountInCents == input.Amount.Value { // A check for exact match that has higher priority then earlie transfers
					log.Infof(c, "Found outstanding transfer with exact amount match: %v", outstandingTransferID)
					assignedValue = input.Amount.Value
					returnToTransferIDs = []int64{outstandingTransferID}
					break
				}
				if assignedValue < input.Amount.Value { // Do not break so we check all outstanding transfers for exact match
					returnToTransferIDs = append(returnToTransferIDs, outstandingTransferID)
					assignedValue += outstandingTransfer.AmountInCentsOutstanding
				}
				outstandingRightDirection += 1
			}
		}
		if input.IsReturn && assignedValue < input.Amount.Value {
			m := fmt.Sprintf(
				"There are not enough outstanding transfers to return %v. All outstading count: %v, Right direction: %v, Assigned amount: %v. Could be data integrity issue.",
				input.Amount, len(outstandingTransfers), outstandingRightDirection, assignedValue,
			)
			log.Errorf(c, m)
		}
	}
	return
}

func (transferFacade transferFacade) createTransferWithinTransaction(
	c context.Context, input createTransferInput, returnToTransferIDs []int64,
) (
	output createTransferOutput, err error,
) {
	log.Debugf(c, "createTransferWithinTransaction(input=%v, returnToTransferIDs=%v)", input, returnToTransferIDs)

	input.Validate()
	//if len(returnToTransferIDs) > 0 && !input.IsReturn { // TODO: It's OK to have returns without isReturn=true
	//	panic("len(returnToTransferIDs) > 0 && !isReturn")
	//}

	output.From = new(createTransferOutputCounterparty)
	output.To = new(createTransferOutputCounterparty)
	from, to := input.From, input.To

	entities := make([]db.EntityHolder, 0, 4+len(returnToTransferIDs))
	if from.UserID != 0 {
		output.From.User.ID = from.UserID
		entities = append(entities, &output.From.User)
	}
	if to.UserID != 0 {
		output.To.User.ID = to.UserID
		entities = append(entities, &output.To.User)
	}
	if from.ContactID != 0 {
		output.From.Contact.ID = from.ContactID
		entities = append(entities, &output.From.Contact)
	}
	if to.ContactID != 0 {
		output.To.Contact.ID = to.ContactID
		entities = append(entities, &output.To.Contact)
	}

	if err = dal.DB.GetMulti(c, entities); err != nil {
		err = errors.Wrap(err, "Failed to get user & counterparty entities from datastore by keys")
		return
	}
	fromCounterparty, toCounterparty := output.From.Contact, output.To.Contact
	fromUser, toUser := output.From.User, output.To.User

	if from.ContactID != 0 && output.From.Contact.UserID == 0 {
		err = errors.New(fmt.Sprintf("Got bad counterparty entity from DB by id=%d, fromCounterparty.UserID == 0", from.ContactID))
		return
	}
	if to.ContactID != 0 && output.To.Contact.UserID == 0 {
		err = errors.New(fmt.Sprintf("Got bad counterparty entity from DB by id=%d, toCounterparty.UserID == 0", to.ContactID))
		return
	}

	if to.ContactID != 0 && from.ContactID != 0 {
		if fromCounterparty.CounterpartyUserID != toCounterparty.UserID {
			err = errors.New(fmt.Sprintf("fromCounterparty.CounterpartyUserID != toCounterparty.UserID (%d != %d)",
				fromCounterparty.CounterpartyUserID, toCounterparty.UserID))
		}
		if toCounterparty.CounterpartyUserID != fromCounterparty.UserID {
			err = errors.New(fmt.Sprintf("toCounterparty.CounterpartyUserID != fromCounterparty.UserID (%d != %d)",
				toCounterparty.CounterpartyUserID, fromCounterparty.UserID))
		}
		return
	}

	// Check if counterparties are linked and if yes load the missing Contact
	{
		// When: toCounterparty == nil, fromUser == nil,
		if from.ContactID != 0 && fromCounterparty.CounterpartyCounterpartyID != 0 && to.ContactID == 0 {
			// Get toCounterparty and fill to.Contact* fields
			if toCounterparty, err = dal.Contact.GetContactByID(c, fromCounterparty.CounterpartyCounterpartyID); err != nil {
				err = errors.Wrap(err, "Failed to get 'To' counterparty by 'fromCounterparty.CounterpartyCounterpartyID'")
				return
			}
			output.To.Contact = toCounterparty
			log.Debugf(c, "Got toCounterparty id=%d: %v", toCounterparty.ID, toCounterparty.ContactEntity)
			to.ContactID = toCounterparty.ID
			to.ContactName = toCounterparty.FullName()
			from.UserID = toCounterparty.UserID
			entities = append(entities, &toCounterparty)
		}
		if to.ContactID != 0 && toCounterparty.CounterpartyCounterpartyID != 0 && from.ContactID == 0 {
			if fromCounterparty, err = dal.Contact.GetContactByID(c, toCounterparty.CounterpartyCounterpartyID); err != nil {
				err = errors.Wrapf(err, "Failed to get 'From' counterparty by 'toCounterparty.CounterpartyCounterpartyID' == %d", fromCounterparty.CounterpartyCounterpartyID)
				return
			}
			output.From.Contact = fromCounterparty
			log.Debugf(c, "Got fromCounterparty id=%d: %v", fromCounterparty.ID, fromCounterparty.ContactEntity)
			from.ContactID = fromCounterparty.ID
			from.ContactName = fromCounterparty.FullName()
			to.UserID = fromCounterparty.UserID
			entities = append(entities, &fromCounterparty)
		}
	}

	// In case if we just loaded above missing counterparty we need to check for missing user
	{
		loadUserIfNeeded := func(who string, userID int64, appUser models.AppUser) (models.AppUser, models.AppUser, error) {
			log.Debugf(c, "%v.UserID: %d, %vUser.AppUserEntity: %v", who, userID, who, appUser.AppUserEntity)
			if userID != 0 {
				if appUser.AppUserEntity == nil {
					if appUser, err = dal.User.GetUserByID(c, userID); err != nil {
						err = errors.Wrap(err, fmt.Sprintf("Failed to get %vUser for linked counterparty", who))
						return appUser, appUser, err
					}
					entities = append(entities, &appUser)
				} else if userID != appUser.ID {
					panic("userID != appUser.ID")
				}
			}
			return appUser, appUser, err
		}

		if fromUser, output.From.User, err = loadUserIfNeeded("from", from.UserID, fromUser); err != nil {
			return
		}
		if toUser, output.To.User, err = loadUserIfNeeded("to", to.UserID, toUser); err != nil {
			return
		}
	}

	transferEntity := models.NewTransferEntity(input.CreatorUserID, input.IsReturn, input.Amount, input.From, input.To)
	output.Transfer.TransferEntity = transferEntity
	input.Source.PopulateTransfer(transferEntity)

	var (
		transferIDsToDiscard []int64
	)

	// For returns to specific transfers
	if len(returnToTransferIDs) > 0 {
		transferEntity.ReturnToTransferIDs = returnToTransferIDs
		returnToTransfers := make([]db.EntityHolder, len(returnToTransferIDs))
		for i, returnToTransferID := range returnToTransferIDs {
			returnToTransfers[i] = &models.Transfer{ID: returnToTransferID, TransferEntity: new(models.TransferEntity)}
		}
		if err = dal.DB.GetMulti(c, returnToTransfers); err != nil {
			err = errors.Wrapf(err, "Failed to load returnToTransfers by keys (%v)", returnToTransferIDs)
			return
		}
		log.Debugf(c, "Loaded %d returnToTransfers by keys", len(returnToTransfers))
		amountToAssign := input.Amount.Value
		assignedToExistingTransfers := false
		var returnedAmount decimal.Decimal64p2
		for _, transferEntityHolder := range returnToTransfers {
			returnToTransfer := transferEntityHolder.(*models.Transfer)
			if !returnToTransfer.IsOutstanding {
				log.Warningf(c, "Transfer(%v).IsOutstanding: false", returnToTransfer.ID)
				continue
			}
			if returnToTransfer.AmountInCentsOutstanding <= 0 {
				log.Warningf(c, "Transfer(%v).AmountInCentsOutstanding:%v <= 0", returnToTransfer.ID, returnToTransfer.AmountInCentsOutstanding)
				continue
			}
			if amountToAssign < returnToTransfer.AmountInCentsOutstanding {
				returnToTransfer.AmountInCentsReturned += amountToAssign
				returnToTransfer.AmountInCentsOutstanding -= amountToAssign
				returnedAmount += amountToAssign
				amountToAssign = 0
			} else {
				amountToAssign -= returnToTransfer.AmountInCentsOutstanding
				returnedAmount += returnToTransfer.AmountInCentsOutstanding
				returnToTransfer.AmountInCentsReturned += returnToTransfer.AmountInCentsOutstanding
				returnToTransfer.AmountInCentsOutstanding = 0
				returnToTransfer.IsOutstanding = false
				transferIDsToDiscard = append(transferIDsToDiscard, returnToTransfer.ID)
			}
			assignedToExistingTransfers = true
			entities = append(entities, returnToTransfer) // TODO: Potentially can exceed max number of entities in GAE transaction

			if transferEntity.CreatorUserID == returnToTransfer.CreatorUserID && transferEntity.Direction() == returnToTransfer.Direction() {
				panic(fmt.Sprintf(
					"transfer.CreatorUserID == returnToTransfer.CreatorUserID && transfer.Direction == returnToTransfer.Direction, userID=%v, direction=%v",
					transferEntity.CreatorUserID, transferEntity.Direction()))
			}

			if transferEntity.CreatorUserID == returnToTransfer.Counterparty().UserID && transferEntity.Direction() != returnToTransfer.Direction() {
				panic(fmt.Sprintf(
					"transfer.CreatorUserID == returnToTransfer.CounterpartyUserID && transfer.Direction=%v != returnToTransfer.Direction=%v, userID=%v",
					transferEntity.Direction(), returnToTransfer.Direction(), transferEntity.CreatorUserID))
			}

			if amountToAssign == 0 {
				break
			}
		}
		if assignedToExistingTransfers {
			if returnedAmount > 0 {
				if returnedAmount > input.Amount.Value {
					panic("returnedAmount > input.Amount.Value")
				}
				if returnedAmount == input.Amount.Value && !transferEntity.IsReturn {
					transferEntity.IsReturn = true
					transferEntity.AmountInCentsOutstanding = 0
					transferEntity.AmountInCentsReturned = 0
					log.Debugf(c, "Transfer marked IsReturn=true as it's amount less or equal to outstanding debt(s)")
				}
				if returnedAmount != input.Amount.Value {
					transferEntity.AmountInCentsOutstanding = input.Amount.Value - returnedAmount
					transferEntity.AmountInCentsReturned = returnedAmount
				}
			}
			if output.From.User.ID != 0 {
				dal.User.DelayUpdateUserHasDueTransfers(c, output.From.User.ID)
			}
			if output.To.User.ID != 0 {
				dal.User.DelayUpdateUserHasDueTransfers(c, output.To.User.ID)
			}
		}
	}

	if !input.DueOn.IsZero() {
		transferEntity.DtDueOn = input.DueOn
		if from.UserID != 0 {
			output.From.User.HasDueTransfers = true
		}
		if to.UserID != 0 {
			output.To.User.HasDueTransfers = true
		}
	}

	// Set from & to names if needed
	{
		fixUserName := func(counterparty *models.TransferCounterpartyInfo, user models.AppUser) {
			if counterparty.UserID != 0 && counterparty.UserName == "" {
				counterparty.UserName = user.FullName()
			}
		}
		fixUserName(input.From, output.From.User)
		fixUserName(input.To, output.To.User)

		fixContactName := func(counterparty *models.TransferCounterpartyInfo, contact models.Contact) {
			if counterparty.ContactID != 0 && counterparty.ContactName == "" {
				counterparty.ContactName = contact.FullName()
			}
		}
		fixContactName(input.From, output.From.Contact)
		fixContactName(input.To, output.To.Contact)
	}

	log.Debugf(c, "from: %v", input.From)
	log.Debugf(c, "to: %v", input.To)

	log.Debugf(c, "transferEntity before insert: %+v", transferEntity)
	if output.Transfer, err = dal.Transfer.InsertTransfer(c, transferEntity); err != nil {
		err = errors.Wrap(err, "Failed to save transfer entity")
		return
	}

	transfer := output.Transfer

	if output.Transfer.ID == 0 {
		panic(fmt.Sprintf("Can't proceed creating transfer as InsertTransfer() returned transfer.ID == 0, err: %v", err))
	}

	log.Infof(c, "Transfer inserted to DB with ID=%d, %+v", output.Transfer.ID, transfer.TransferEntity)

	// Update user and counterparty entities with transfer info
	{
		if output.From.User.ID != 0 {
			transferFacade.updateUserAndCounterpartyWithTransferInfo(c, input.Amount, output.Transfer, output.From.User, output.To.Contact, models.TransferDirectionUser2Counterparty)
		}
		if output.To.User.ID != 0 {
			transferFacade.updateUserAndCounterpartyWithTransferInfo(c, input.Amount, output.Transfer, output.To.User, output.From.Contact, models.TransferDirectionCounterparty2User)
		}
	}

	if err = dal.DB.UpdateMulti(c, entities); err != nil {
		err = errors.Wrap(err, "Failed to update entities")
		return
	}

	if len(transferIDsToDiscard) > 0 {
		if err = dal.Reminder.DelayDiscardReminders(c, transferIDsToDiscard, transfer.ID); err != nil {
			err = errors.Wrapf(err, "Failed to delay task to discard reminders (transferIDsToDiscard=%v)", len(transferIDsToDiscard), transferIDsToDiscard)
			return
		}
	}

	if output.Transfer.Counterparty().UserID != 0 {
		if err = dal.Receipt.DelaySendReceiptToCounterpartyByTelegram(c, input.Env, transfer.ID, transfer.Counterparty().UserID); err != nil {
			// TODO: Send by any available channel
			err = errors.Wrap(err, "Failed to delay sending receipt to counterpartyEntity by Telegram")
			return
		}
	} else {
		log.Debugf(c, "No receipt to counterpartyEntity: [%v]", transfer.Counterparty().ContactName)
	}

	if err = dal.Reminder.DelayCreateReminderForTransferCreator(c, transfer.ID); err != nil {
		err = errors.Wrap(err, "Failed to delay reminder creation for creator")
		return
	}

	log.Debugf(c, "createTransferWithinTransaction(): transferID=%v", transfer.ID)
	return
}

func (_ transferFacade) updateUserAndCounterpartyWithTransferInfo(
	c context.Context,
	amount models.Amount,
	transfer models.Transfer,
	user models.AppUser,
	counterparty models.Contact,
	direction models.TransferDirection,
) (err error) {
	if user.ID != counterparty.UserID {
		panic(fmt.Sprintf("user.ID != counterparty.UserID (%d != %d)", user.ID, counterparty.UserID))
	}

	var updateBalance = func(curr models.Currency, val decimal.Decimal64p2, user models.AppUser, cp models.Contact) (err error) {
		log.Debugf(c, "Updating balance with [%v %v] for user #%d, counterparty #%d", val, curr, user.ID, cp.ID)
		if user.ID != cp.UserID {
			panic("user.ID != cp.UserID")
		}
		var balance models.Balance
		if balance, err = cp.Add2Balance(curr, val); err != nil {
			err = errors.Wrapf(err, "Failed to add (%v %v) to balance for counterparty #%d", curr, val, cp.ID)
			return
		} else {
			cp.CountOfTransfers += 1
			cpBalance, _ := cp.Balance()
			log.Debugf(c, "Updated balance to %v | %v for counterparty #%d", balance, cpBalance, cp.ID)
		}
		if balance, err = user.Add2Balance(curr, val); err != nil {
			err = errors.Wrapf(err, "Failed to add (%v %v) to balance for user #%d", curr, val, user.ID)
			return
		} else {
			user.CountOfTransfers += 1
			userBalance, _ := user.Balance()
			log.Debugf(c, "Updated balance to %v | %v for user #%d", balance, userBalance, user.ID)
		}

		log.Debugf(c, "user.ContactsJson (before): %v", user.ContactsJson)
		userCounterpartiesChanged := user.AddOrUpdateContact(cp)
		log.Debugf(c, "user.ContactsJson (changed=%v): %v", userCounterpartiesChanged, user.ContactsJson)
		return
	}

	counterparty.LastTransferID = transfer.ID
	counterparty.LastTransferAt = transfer.DtCreated

	user.LastTransferID = transfer.ID
	user.LastTransferAt = transfer.DtCreated
	user.SetLastCurrency(transfer.Currency)

	var amountValue decimal.Decimal64p2
	switch direction {
	case models.TransferDirectionUser2Counterparty:
		amountValue = amount.Value * USER_BALANCE_INCREASED
	case models.TransferDirectionCounterparty2User:
		amountValue = amount.Value * USER_BALANCE_DECREASED
	default:
		panic("Unknown direciton: " + string(direction))
	}
	if err = updateBalance(amount.Currency, amountValue, user, counterparty); err != nil {
		return
	}
	return
}
