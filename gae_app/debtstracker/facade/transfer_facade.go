package facade

import (
	"bytes"
	"fmt"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/crediterra/money"
	"github.com/sanity-io/litter"
	"github.com/strongo/app"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"github.com/strongo/slices"
)

const (
	userBalanceIncreased = 1
	userBalanceDecreased = -1
)

type TransfersFacade interface {
	GetTransferByID(c context.Context, id int) (transfer models.Transfer, err error)
	SaveTransfer(c context.Context, transfer models.Transfer) error
	CreateTransfer(c context.Context, input createTransferInput) (output createTransferOutput, err error)
	UpdateTransferOnReturn(c context.Context, returnTransfer, transfer models.Transfer, returnedAmount decimal.Decimal64p2) (err error)
}

var (
	ErrNotImplemented                      = errors.New("not implemented yet")
	ErrDebtAlreadyReturned                 = errors.New("This debt already has been returned")
	ErrPartialReturnGreaterThenOutstanding = errors.New("An attempt to do partial return for amount greater then outstanding")
	//
	ErrNoOutstandingTransfers                                       = errors.New("no outstanding transfers")
	ErrAttemptToCreateDebtWithInterestAffectingOutstandingTransfers = errors.New("You are trying to create a debt with interest that will affect outstanding transfers. Please close them first.")
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

var Transfers TransfersFacade = transferFacade{}

func (transferFacade) SaveTransfer(c context.Context, transfer models.Transfer) error {
	return dtdal.DB.Update(c, &transfer)
}

type createTransferInput struct {
	Env                strongo.Environment // TODO: I believe we don't need this
	Source             dtdal.TransferSource
	CreatorUser        models.AppUser
	BillID             string
	IsReturn           bool
	ReturnToTransferID int64
	// direction models.TransferDirection,
	// creatorInfo models.TransferCounterpartyInfo,
	From, To *models.TransferCounterpartyInfo
	Amount   money.Amount
	DueOn    time.Time
	Interest models.TransferInterest
}

func (input createTransferInput) Direction() (direction models.TransferDirection) {
	if input.CreatorUser.ID == 0 {
		panic("createTransferInput.CreatorUserID == 0")
	}
	switch input.CreatorUser.ID {
	case input.From.UserID:
		return models.TransferDirectionUser2Counterparty
	case input.To.UserID:
		return models.TransferDirectionCounterparty2User
	default:
		if input.BillID == "" {
			panic("Not able to detect direction")
		}
		return models.TransferDirection3dParty
	}
}

func (input createTransferInput) CreatorContactID() int64 {
	switch input.CreatorUser.ID {
	case input.From.UserID:
		return input.To.ContactID
	case input.To.UserID:
		return input.From.ContactID
	}
	panic("Can't get creator's contact ID as it's a 3d-party transfer")
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

func (input createTransferInput) Validate() {
	if input.Source == nil {
		panic("source == nil")
	}
	if input.CreatorUser.ID == 0 {
		panic("creatorUser.ID == 0")
	}
	if input.CreatorUser.AppUserEntity == nil {
		panic("creatorUser.AppUserEntity == nil")
	}
	if input.Amount.Value <= 0 {
		panic("amount.Value <= 0")
	}
	if input.From == nil {
		panic("from == nil")
	}
	if input.To == nil {
		panic("to == nil")
	}

	if (input.From.ContactID == 0 && input.To.ContactID == 0) || (input.From.UserID == 0 && input.To.UserID == 0) {
		panic("(from.ContactID == 0  && to.ContactID == 0) || (from.UserID == 0 && to.UserID == 0)")
	}
	if input.From.UserID != 0 && input.To.ContactID == 0 && input.To.UserID == 0 {
		panic("from.UserID != 0 && to.ContactID == 0 && to.UserID == 0")
	}
	if input.To.UserID != 0 && input.From.ContactID == 0 && input.From.UserID == 0 {
		panic("to.UserID != 0 && from.ContactID == 0 && from.UserID == 0")
	}

	if input.From.UserID == input.To.UserID {
		if input.From.UserID == 0 && input.To.UserID == 0 {
			if input.From.ContactID == 0 {
				panic("from.UserID == 0 && to.UserID == 0 && from.ContactID == 0")
			}
			if input.To.ContactID == 0 {
				panic("from.UserID == 0 && to.UserID == 0 && to.ContactID == 0")
			}
		} else {
			panic("from.UserID == to.UserID")
		}
	}
	switch input.CreatorUser.ID {
	case input.From.UserID:
		if input.To.ContactID == 0 {
			panic("creatorUserID == from.UserID && to.ContactID == 0")
		}
	case input.To.UserID:
		if input.From.ContactID == 0 {
			panic("creatorUserID == from.UserID && from.ContactID == 0")
		}
	default:
		if input.From.ContactID == 0 {
			panic("3d party transfer and from.ContactID == 0")
		}
		if input.To.ContactID == 0 {
			panic("3d party transfer and to.ContactID == 0")
		}
	}
}

func (input createTransferInput) String() string {
	return fmt.Sprintf("CreatorUserID=%d, IsReturn=%v, ReturnToTransferID=%d, Amount=%v, From=%v, To=%v, DueOn=%v",
		input.CreatorUser.ID, input.IsReturn, input.ReturnToTransferID, input.Amount, input.From, input.To, input.DueOn)
}

func NewTransferInput(
	env strongo.Environment,
	source dtdal.TransferSource,
	creatorUser models.AppUser,
	billID string,
	isReturn bool, returnToTransferID int64,
	from, to *models.TransferCounterpartyInfo,
	amount money.Amount,
	dueOn time.Time,
	transferInterest models.TransferInterest,
) (input createTransferInput) {
	// All checks are in the input.Validate()
	input = createTransferInput{
		Env:                env,
		Source:             source,
		CreatorUser:        creatorUser,
		BillID:             billID,
		IsReturn:           isReturn,
		ReturnToTransferID: returnToTransferID,
		From:               from,
		To:                 to,
		Amount:             amount,
		DueOn:              dueOn,
		Interest:           transferInterest,
	}
	input.Validate()
	return
}

func (transferFacade transferFacade) CreateTransfer(c context.Context, input createTransferInput) (
	output createTransferOutput, err error,
) {
	now := time.Now()

	log.Infof(c, "CreateTransfer(input=%v)", input)

	var returnToTransferIDs []int64

	if input.ReturnToTransferID == 0 {
		log.Debugf(c, "input.ReturnToTransferID == 0")
		contacts := input.CreatorUser.Contacts()
		creatorContactID := input.CreatorContactID()
		if creatorContactID == 0 {
			panic(errors.WithMessage(err, "3d party transfers are not implemented yet"))
		}
		log.Debugf(c, "creatorContactID=%v, contacts: %+v", creatorContactID, contacts)
		var creatorContact models.Contact
		verifyUserContactJson := func() (contactJsonFound bool) {
			for _, contact := range contacts {
				if contact.ID == creatorContactID {
					contactBalance := contact.Balance()
					if v, ok := contactBalance[input.Amount.Currency]; !ok || v == 0 {
						log.Debugf(c, "No need to check for outstanding transfers as contacts balance is 0")
					} else {
						if input.Interest.HasInterest() {
							if d := input.Direction(); d == models.TransferDirectionUser2Counterparty && v < 0 || d == models.TransferDirectionCounterparty2User && v > 0 {
								err = ErrAttemptToCreateDebtWithInterestAffectingOutstandingTransfers
								return
							}
						}
						if returnToTransferIDs, err = transferFacade.checkOutstandingTransfersForReturns(c, now, input); err != nil {
							return
						}
					}
					contactJsonFound = true
					return
				}
			}
			return
		}
		if contactJsonFound := verifyUserContactJson(); contactJsonFound {
			goto contactFound
		}
		// If contact not found in user's JSON try to recover from DB record
		if creatorContact, err = GetContactByID(c, creatorContactID); err != nil {
			return
		}

		log.Warningf(c, "data integrity issue: contact found by ID in database but is missing in user's JSON: creatorContactID=%v, creatorContact.UserID=%v, user.ID=%v, user.ContactsJsonActive: %v",
			creatorContactID, creatorContact.UserID, input.CreatorUser.ID, input.CreatorUser.ContactsJsonActive)

		if creatorContact.UserID != input.CreatorUser.ID {
			err = fmt.Errorf("creatorContact.UserID != input.CreatorUser.ID: %v != %v", creatorContact.UserID, input.CreatorUser.ID)
			return
		}

		if _, changed := input.CreatorUser.AddOrUpdateContact(creatorContact); changed {
			contacts = input.CreatorUser.Contacts()
		}
		if contactJsonFound := verifyUserContactJson(); contactJsonFound {
			goto contactFound
		}
		if err == nil {
			err = fmt.Errorf("user contact not found by ID=%v, contacts: %v", creatorContactID, litter.Sdump(contacts))
		}
		return
	contactFound:
	} else if !input.IsReturn {
		panic("ReturnToTransferID != 0 && !IsReturn")
	}
	if input.ReturnToTransferID != 0 {
		var transferToReturn models.Transfer
		if transferToReturn, err = Transfers.GetTransferByID(c, input.ReturnToTransferID); err != nil {
			err = errors.Wrapf(err, "Failed to get returnToTransferID=%v", input.ReturnToTransferID)
			return
		}

		if transferToReturn.Currency != input.Amount.Currency {
			panic("transferToReturn.Currency != amount.Currency")
		}

		if transferToReturn.GetOutstandingValue(now) == 0 {
			// When the transfer has been already returned
			err = ErrDebtAlreadyReturned
			return
		}

		if input.Amount.Value > transferToReturn.GetOutstandingValue(now) {
			log.Debugf(c, "amount.Value:%v > transferToReturn.GetOutstandingValue(now):%v", input.Amount.Value, transferToReturn.GetOutstandingValue(now))
			if input.Amount.Value == transferToReturn.AmountInCents {
				// For situations when a transfer was partially returned but user wants to mark it as fully returned.
				log.Debugf(c, "amount.Value (%v) == transferToReturn.AmountInCents (%v)", input.Amount.Value, transferToReturn.AmountInCents)
				input.Amount.Value = transferToReturn.GetOutstandingValue(now)
				log.Debugf(c, "Updated amount.Value: %v", input.Amount.Value)
			} else {
				err = ErrPartialReturnGreaterThenOutstanding
				return
			}
		} else if input.Amount.Value < transferToReturn.GetOutstandingValue(now) {
			log.Debugf(c, "input.Amount.Value < transferToReturn.GetOutstandingValue(now)")
		}

		returnToTransferIDs = append(returnToTransferIDs, input.ReturnToTransferID)
		output.ReturnedTransfers = append(output.ReturnedTransfers, transferToReturn)
	}

	if err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
		output, err = transferFacade.createTransferWithinTransaction(c, now, input, returnToTransferIDs)
		return err
	}, dtdal.CrossGroupTransaction); err != nil {
		return
	}

	output.Validate()

	return
}

func (transferFacade transferFacade) checkOutstandingTransfersForReturns(c context.Context, now time.Time, input createTransferInput) (returnToTransferIDs []int64, err error) {
	log.Debugf(c, "transferFacade.checkOutstandingTransfersForReturns()")
	var (
		outstandingTransfers []models.Transfer
	)

	creatorUserID := input.CreatorUser.ID
	creatorContactID := input.CreatorContactID()

	reversedDirection := input.Direction().Reverse()
	outstandingTransfers, err = dtdal.Transfer.LoadOutstandingTransfers(c, now, creatorUserID, creatorContactID, input.Amount.Currency, reversedDirection)
	if err != nil {
		err = errors.WithMessage(err, "failed to load outstanding transfers")
		return
	}
	if input.IsReturn && len(outstandingTransfers) == 0 {
		err = ErrNoOutstandingTransfers
		return
	}

	log.Debugf(c, "facade.checkOutstandingTransfersForReturns() => dtdal.Transfer.LoadOutstandingTransfers(userID=%v, currency=%v) => %d transfers", input.CreatorUser.ID, input.Amount.Currency, len(outstandingTransfers))

	if outstandingTransfersCount := len(outstandingTransfers); outstandingTransfersCount > 0 { // Assign the return to specific transfers
		var (
			assignedValue             decimal.Decimal64p2
			outstandingRightDirection int
		)
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "%v outstanding transfers\n", outstandingTransfersCount)
		for i, outstandingTransfer := range outstandingTransfers {
			fmt.Fprintf(buf, "\t[%v]: %v", i, litter.Sdump(outstandingTransfer))
			outstandingTransferID := outstandingTransfers[i].ID
			outstandingValue := outstandingTransfer.GetOutstandingValue(now)
			if outstandingValue == input.Amount.Value { // A check for exact match that has higher priority then earlie transfers
				log.Infof(c, " - found outstanding transfer %v with exact amount match: %v", outstandingTransfer.ID, outstandingValue)
				assignedValue = input.Amount.Value
				returnToTransferIDs = []int64{outstandingTransferID}
				break
			}
			if assignedValue < input.Amount.Value { // Do not break so we check all outstanding transfers for exact match
				returnToTransferIDs = append(returnToTransferIDs, outstandingTransferID)
				assignedValue += outstandingValue
			}
			outstandingRightDirection += 1
			buf.WriteString("\n")
		}
		log.Debugf(c, buf.String())
		if input.IsReturn && assignedValue < input.Amount.Value {
			log.Warningf(c,
				"There are not enough outstanding transfers to return %v. All outstanding count: %v, Right direction: %v, Assigned amount: %v. Could be data integrity issue.",
				input.Amount, len(outstandingTransfers), outstandingRightDirection, assignedValue,
			)
		}
	}
	return
}

func (transferFacade transferFacade) createTransferWithinTransaction(
	c context.Context, dtCreated time.Time, input createTransferInput, returnToTransferIDs []int64,
) (
	output createTransferOutput, err error,
) {
	log.Debugf(c, "createTransferWithinTransaction(input=%v, returnToTransferIDs=%v)", input, returnToTransferIDs)

	input.Validate()
	// if len(returnToTransferIDs) > 0 && !input.IsReturn { // TODO: It's OK to have transfers without isReturn=true
	// 	panic("len(returnToTransferIDs) > 0 && !isReturn")
	// }

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

	if err = dtdal.DB.GetMulti(c, entities); err != nil {
		err = errors.WithMessage(err, "failed to get user & counterparty entities from datastore by keys")
		return
	}
	fromContact, toContact := output.From.Contact, output.To.Contact
	fromUser, toUser := output.From.User, output.To.User

	if from.ContactID != 0 && output.From.Contact.UserID == 0 {
		err = fmt.Errorf("got bad counterparty entity from DB by id=%d, fromCounterparty.UserID == 0", from.ContactID)
		return
	}

	if to.ContactID != 0 && output.To.Contact.UserID == 0 {
		err = fmt.Errorf("got bad counterparty entity from DB by id=%d, toCounterparty.UserID == 0", to.ContactID)
		return
	}

	if to.ContactID != 0 && from.ContactID != 0 {
		if fromContact.CounterpartyUserID != toContact.UserID {
			err = fmt.Errorf("fromCounterparty.CounterpartyUserID != toCounterparty.UserID (%d != %d)",
				fromContact.CounterpartyUserID, toContact.UserID)
		}
		if toContact.CounterpartyUserID != fromContact.UserID {
			err = fmt.Errorf("toCounterparty.CounterpartyUserID != fromCounterparty.UserID (%d != %d)",
				toContact.CounterpartyUserID, fromContact.UserID)
		}
		return
	}

	// Check if counterparties are linked and if yes load the missing Contact
	{
		link := func(sideName, countersideName string, side, counterside *models.TransferCounterpartyInfo, sideContact models.Contact) (countersideContact models.Contact, err error) {
			log.Debugf(c, "link(%v=%v, %v=%v, %vContact=%v)", sideName, side, countersideName, counterside, sideName, sideContact)
			if side.ContactID != 0 && sideContact.CounterpartyCounterpartyID != 0 && counterside.ContactID == 0 {
				if countersideContact, err = GetContactByID(c, sideContact.CounterpartyCounterpartyID); err != nil {
					err = errors.WithMessage(err, "Failed to get counterparty by 'fromCounterparty.CounterpartyCounterpartyID'")
					return
				}
				counterside.ContactID = countersideContact.ID
				counterside.ContactName = countersideContact.FullName()
				side.UserID = countersideContact.UserID
				entities = append(entities, &countersideContact)
			}
			return
		}

		var linkedContact models.Contact // TODO: This smells
		if linkedContact, err = link("from", "to", from, to, fromContact); err != nil {
			return
		} else if linkedContact.ContactEntity != nil {
			toContact = linkedContact
			output.To.Contact = linkedContact
		}

		log.Debugf(c, "toContact: %v", toContact.ContactEntity == nil)
		if linkedContact, err = link("to", "from", to, from, toContact); err != nil {
			return
		} else if linkedContact.ContactEntity != nil {
			fromContact = linkedContact
			output.From.Contact = fromContact
		}

		// //// When: toCounterparty == nil, fromUser == nil,
		// if from.ContactID != 0 && fromContact.CounterpartyCounterpartyID != 0 && to.ContactID == 0 {
		// 	// Get toCounterparty and fill to.Contact* fields
		// 	if toContact, err = GetContactByID(c, fromContact.CounterpartyCounterpartyID); err != nil {
		// 		err = errors.WithMessage(err, "Failed to get 'To' counterparty by 'fromCounterparty.CounterpartyCounterpartyID'")
		// 		return
		// 	}
		// 	output.To.Contact = toContact
		// 	log.Debugf(c, "Got toContact id=%d: %v", toContact.ID, toContact.ContactEntity)
		// 	to.ContactID = toContact.ID
		// 	to.ContactName = toContact.GetFullName()
		// 	from.UserID = toContact.UserID
		// 	entities = append(entities, &toContact)
		// }
		// if to.ContactID != 0 && toCounterparty.CounterpartyCounterpartyID != 0 && from.ContactID == 0 {
		// 	if fromCounterparty, err = GetContactByID(c, toCounterparty.CounterpartyCounterpartyID); err != nil {
		// 		err = errors.WithMessage(err, fmt.Sprintf("Failed to get 'From' counterparty by 'toCounterparty.CounterpartyCounterpartyID' == %d", fromCounterparty.CounterpartyCounterpartyID))
		// 		return
		// 	}
		// 	output.From.Contact = fromCounterparty
		// 	log.Debugf(c, "Got fromCounterparty id=%d: %v", fromCounterparty.ID, fromCounterparty.ContactEntity)
		// 	from.ContactID = fromCounterparty.ID
		// 	from.ContactName = fromCounterparty.GetFullName()
		// 	to.UserID = fromCounterparty.UserID
		// 	entities = append(entities, &fromCounterparty)
		// }
	}

	// In case if we just loaded above missing counterparty we need to check for missing user
	{
		loadUserIfNeeded := func(who string, userID int64, appUser models.AppUser) (models.AppUser, models.AppUser, error) {
			log.Debugf(c, "%v.UserID: %d, %vUser.AppUserEntity: %v", who, userID, who, appUser.AppUserEntity)
			if userID != 0 {
				if appUser.AppUserEntity == nil {
					if appUser, err = User.GetUserByID(c, userID); err != nil {
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

	transferEntity := models.NewTransferEntity(input.CreatorUser.ID, input.IsReturn, input.Amount, input.From, input.To)
	transferEntity.DtCreated = dtCreated
	output.Transfer.TransferEntity = transferEntity
	input.Source.PopulateTransfer(transferEntity)
	transferEntity.TransferInterest = input.Interest

	type TransferReturnInfo struct {
		Transfer       models.Transfer
		ReturnedAmount decimal.Decimal64p2
	}

	var (
		transferReturnInfos             = make([]TransferReturnInfo, 0, len(returnToTransferIDs))
		returnedValue, returnedInterest decimal.Decimal64p2
		closedTransferIDs               []int64
	)

	// For transfers to specific transfers
	if len(returnToTransferIDs) > 0 {
		transferEntity.ReturnToTransferIDs = returnToTransferIDs
		returnToTransfers := make([]db.EntityHolder, len(returnToTransferIDs))
		for i, returnToTransferID := range returnToTransferIDs {
			returnToTransfers[i] = &models.Transfer{IntegerID: db.NewIntID(returnToTransferID), TransferEntity: new(models.TransferEntity)}
		}
		if err = dtdal.DB.GetMulti(c, returnToTransfers); err != nil { // TODO: This can exceed limit on TX entity groups
			err = errors.WithMessage(err, fmt.Sprintf("failed to load returnToTransfers by keys (%v)", returnToTransferIDs))
			return
		}
		log.Debugf(c, "Loaded %d returnToTransfers by keys", len(returnToTransfers))
		amountToAssign := input.Amount.Value
		assignedToExistingTransfers := false
		for _, transferEntityHolder := range returnToTransfers {
			returnToTransfer := transferEntityHolder.(*models.Transfer)
			returnToTransferOutstandingValue := returnToTransfer.GetOutstandingValue(dtCreated)
			if !returnToTransfer.IsOutstanding {
				log.Warningf(c, "Transfer(%v).IsOutstanding: false, returnToTransferOutstandingValue: %v", returnToTransfer.ID, returnToTransferOutstandingValue)
				continue
			} else if returnToTransferOutstandingValue == 0 {
				log.Warningf(c, "Transfer(%v) => returnToTransferOutstandingValue == 0", returnToTransfer.ID, returnToTransferOutstandingValue)
				continue
			} else if returnToTransferOutstandingValue < 0 {
				panic(fmt.Sprintf("Transfer(%v) => returnToTransferOutstandingValue:%d <= 0", returnToTransfer.ID, returnToTransferOutstandingValue))
			}
			var amountReturnedToTransfer decimal.Decimal64p2
			if amountToAssign < returnToTransferOutstandingValue {
				amountReturnedToTransfer = amountToAssign
			} else {
				amountReturnedToTransfer = returnToTransferOutstandingValue
			}
			interestReturnedToTransfer := returnToTransfer.GetInterestValue(dtCreated)
			if interestReturnedToTransfer > 0 {
				if interestReturnedToTransfer > amountReturnedToTransfer {
					interestReturnedToTransfer = amountReturnedToTransfer
				}
				returnedInterest += interestReturnedToTransfer
			}
			transferReturnInfos = append(transferReturnInfos, TransferReturnInfo{Transfer: *returnToTransfer, ReturnedAmount: amountReturnedToTransfer})
			amountToAssign -= amountReturnedToTransfer
			returnedValue += amountReturnedToTransfer

			if err = transferEntity.AddReturn(models.TransferReturnJson{
				TransferID: returnToTransfer.ID,
				Amount:     amountReturnedToTransfer,
				Time:       returnToTransfer.DtCreated,
			}); err != nil {
				return
			}

			assignedToExistingTransfers = true
			entities = append(entities, returnToTransfer) // TODO: Potentially can exceed max number of entities in GAE transaction

			if transferEntity.CreatorUserID == returnToTransfer.CreatorUserID && transferEntity.Direction() == returnToTransfer.Direction() {
				panic(fmt.Sprintf(
					"transfer.CreatorUserID == returnToTransfer.CreatorUserID && transfer.Direction == returnToTransfer.Direction, userID=%v, direction=%v, returnToTransfer=%v",
					transferEntity.CreatorUserID, transferEntity.Direction(), returnToTransfer.ID))
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
			if returnedValue > 0 {
				if returnedValue > input.Amount.Value {
					panic("returnedAmount > input.Amount.Value")
				}
				if returnedValue == input.Amount.Value && !transferEntity.IsReturn {
					transferEntity.IsReturn = true
					// transferEntity.AmountInCentsOutstanding = 0
					// transferEntity.AmountInCentsReturned = 0
					log.Debugf(c, "Transfer marked IsReturn=true as it's amount less or equal to outstanding debt(s)")
				}
				// if returnedValue != input.Amount.Value {
				// 	// transferEntity.AmountInCentsOutstanding = input.Amount.Value - returnedAmount
				// 	transferEntity.AmountInCentsReturned = returnedValue
				// }
			}
			if output.From.User.ID != 0 {
				dtdal.User.DelayUpdateUserHasDueTransfers(c, output.From.User.ID)
			}
			if output.To.User.ID != 0 {
				dtdal.User.DelayUpdateUserHasDueTransfers(c, output.To.User.ID)
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
	transferEntity.AmountInCentsInterest = returnedInterest

	// log.Debugf(c, "transferEntity before insert: %v", litter.Sdump(transferEntity))
	if output.Transfer, err = InsertTransfer(c, transferEntity); err != nil {
		err = errors.WithMessage(err, "failed to save transfer entity")
		return
	}

	createdTransfer := output.Transfer

	if output.Transfer.ID == 0 {
		panic(fmt.Sprintf("Can't proceed creating transfer as InsertTransfer() returned transfer.ID == 0, err: %v", err))
	}

	log.Infof(c, "Transfer inserted to DB with ID=%d, %+v", output.Transfer.ID, createdTransfer.TransferEntity)

	if len(transferReturnInfos) > 2 {
		transferReturnUpdates := make([]dtdal.TransferReturnUpdate, len(transferReturnInfos))
		for i, tri := range transferReturnInfos {
			transferReturnUpdates[i] = dtdal.TransferReturnUpdate{TransferID: tri.Transfer.ID, ReturnedAmount: tri.ReturnedAmount}
		}
		dtdal.Transfer.DelayUpdateTransfersOnReturn(c, createdTransfer.ID, transferReturnUpdates)
	} else {
		for _, transferReturnInfo := range transferReturnInfos {
			if err = Transfers.UpdateTransferOnReturn(c, createdTransfer, transferReturnInfo.Transfer, transferReturnInfo.ReturnedAmount); err != nil {
				return
			}
			if !transferReturnInfo.Transfer.IsOutstanding {
				closedTransferIDs = append(closedTransferIDs, transferReturnInfo.Transfer.ID)
			}
		}
	}

	// Update user and counterparty entities with transfer info
	{
		var amountWithoutInterest money.Amount
		if returnedValue > 0 {
			amountWithoutInterest = money.Amount{Currency: input.Amount.Currency, Value: input.Amount.Value - returnedInterest}
		} else if returnedValue < 0 {
			panic(fmt.Sprintf("returnedValue < 0: %v", returnedValue))
		} else {
			amountWithoutInterest = input.Amount
		}

		log.Debugf(c, "closedTransferIDs: %v", closedTransferIDs)

		if output.From.User.ID == output.To.User.ID {
			panic(fmt.Sprintf("output.From.User.ID == output.To.User.ID: %v", output.From.User.ID))
		}
		if output.From.Contact.ID == output.To.Contact.ID {
			panic(fmt.Sprintf("output.From.Contact.ID == output.To.Contact.ID: %v", output.From.Contact.ID))
		}

		if output.From.User.ID != 0 {
			if err = transferFacade.updateUserAndCounterpartyWithTransferInfo(c, amountWithoutInterest, output.Transfer, output.From.User, output.To.Contact, closedTransferIDs); err != nil {
				return
			}
		}
		if output.To.User.ID != 0 {
			if err = transferFacade.updateUserAndCounterpartyWithTransferInfo(c, amountWithoutInterest, output.Transfer, output.To.User, output.From.Contact, closedTransferIDs); err != nil {
				return
			}
		}
	}

	{ // Integrity checks
		checkContacts := func(c1, c2 string, contact models.Contact, user models.AppUser) {
			contacts := user.Contacts()
			contactBalance := contact.Balance()
			for _, c := range contacts {
				if c.ID == contact.ID {
					cBalance := c.Balance()
					for currency, val := range contactBalance {
						if cVal := cBalance[currency]; cVal != val {
							panic(fmt.Sprintf(
								"balance inconsistency for (user=%v&contact=%v VS user=%v&contact=%v) => "+
									"%v: %v != %v\n%v.Balance: %v\n\n%v.Balance: %v",
								contact.UserID, contact.ID, user.ID, c.ID, currency, cVal, val, c1, contactBalance, c2, cBalance))
						}
					}
					return
				}
			}
			panic(fmt.Sprintf("Contact.ID not found in counterparty Contacts(): %v", contact.ID))
		}

		if output.From.User.AppUserEntity != nil {
			checkContacts("to", "from", output.To.Contact, output.From.User)
		}
		if output.To.User.AppUserEntity != nil {
			checkContacts("from", "to", output.From.Contact, output.To.User)
		}
		if output.From.User.AppUserEntity != nil && output.To.User.AppUserEntity != nil {
			currency := output.Transfer.Currency
			fromBalance := output.From.Contact.Balance()[currency]
			toBalance := output.To.Contact.Balance()[currency]
			if fromBalance != -toBalance {
				panic(fmt.Sprintf("from.Contact.Balance != -1*to.Contact.Balance => %v != -1*%v", fromBalance, -toBalance))
			}
		}
	}

	if err = dtdal.DB.UpdateMulti(c, entities); err != nil {
		err = errors.WithMessage(err, "failed to update entities")
		return
	}

	if output.Transfer.Counterparty().UserID != 0 {
		if err = dtdal.Receipt.DelayCreateAndSendReceiptToCounterpartyByTelegram(c, input.Env, createdTransfer.ID, createdTransfer.Counterparty().UserID); err != nil {
			// TODO: Send by any available channel
			err = errors.WithMessage(err, "failed to delay sending receipt to counterpartyEntity by Telegram")
			return
		}
	} else {
		log.Debugf(c, "No receipt to counterpartyEntity: [%v]", createdTransfer.Counterparty().ContactName)
	}

	if createdTransfer.IsOutstanding && dtdal.Reminder != nil { // TODO: check for nil is temporary workaround for unittest
		if err = dtdal.Reminder.DelayCreateReminderForTransferUser(c, createdTransfer.ID, createdTransfer.CreatorUserID); err != nil {
			err = errors.WithMessage(err, "failed to delay reminder creation for creator")
			return
		}
	}

	log.Debugf(c, "createTransferWithinTransaction(): transferID=%v", createdTransfer.ID)
	return
}

func (transferFacade) GetTransferByID(c context.Context, id int64) (transfer models.Transfer, err error) {
	transfer.ID = id
	err = dtdal.DB.Get(c, &transfer)
	return
}

func (transferFacade) updateUserAndCounterpartyWithTransferInfo(
	c context.Context,
	amount money.Amount,
	transfer models.Transfer,
	user models.AppUser,
	contact models.Contact,
	closedTransferIDs []int64,
) (err error) {
	log.Debugf(c, "updateUserAndCounterpartyWithTransferInfo(user=%v, contact=%v)", user, contact)
	if user.ID != contact.UserID {
		panic(fmt.Sprintf("user.ID != contact.UserID (%d != %d)", user.ID, contact.UserID))
	}
	var val decimal.Decimal64p2
	switch user.ID {
	case transfer.From().UserID:
		val = amount.Value * userBalanceIncreased
	case transfer.To().UserID:
		val = amount.Value * userBalanceDecreased
	default:
		panic(fmt.Sprintf("user is not related to transfer: %v", user.ID))
	}
	log.Debugf(c, "Updating balance with [%v %v] for user #%d, contact #%d", val, amount.Currency, user.ID, contact.ID)

	if err = updateContactWithTransferInfo(c, val, transfer, contact, closedTransferIDs); err != nil {
		return
	}
	if err = updateUserWithTransferInfo(c, val, transfer, user, contact, closedTransferIDs); err != nil {
		return
	}
	return
}

func updateUserWithTransferInfo(
	c context.Context,
	val decimal.Decimal64p2,
	// curr money.Currency,
	transfer models.Transfer,
	user models.AppUser,
	contact models.Contact,
	// contact models.Contact,
	closedTransferIDs []int64,
) (err error) {
	user.LastTransferID = transfer.ID
	user.LastTransferAt = transfer.DtCreated
	user.SetLastCurrency(string(transfer.Currency))

	// var updateBalanceAndContactTransfersInfo = func(curr money.Currency, val decimal.Decimal64p2, user models.AppUser, contact models.Contact) (err error) {

	var balance money.Balance
	if balance, err = user.AddToBalance(transfer.Currency, val); err != nil {
		err = errors.WithMessage(err, fmt.Sprintf("failed to add %v=%v to balance for user %v", transfer.Currency, val, user.ID))
		return
	} else {
		user.CountOfTransfers += 1
		userBalance := user.Balance()
		log.Debugf(c, "Updated balance to %v | %v for user #%d", balance, userBalance, user.ID)
	}
	log.Debugf(c, "user.ContactsJsonActive (before): %v\ncontact: %v", user.ContactsJsonActive, litter.Sdump(contact))
	_, userContactsChanged := user.AddOrUpdateContact(contact)
	log.Debugf(c, "user.ContactsJson (changed=%v): %v", userContactsChanged, user.ContactsJsonActive)
	return
}

func updateContactWithTransferInfo(
	c context.Context,
	val decimal.Decimal64p2,
	transfer models.Transfer,
	contact models.Contact,
	closedTransferIDs []int64,
) (err error) {
	contact.LastTransferID = transfer.ID
	contact.LastTransferAt = transfer.DtCreated

	var balance money.Balance
	if balance, err = contact.AddToBalance(transfer.Currency, val); err != nil {
		err = errors.Wrapf(err, "Failed to add (%v %v) to balance for contact #%d", transfer.Currency, val, contact.ID)
		return
	} else {
		contact.CountOfTransfers += 1
		cpBalance := contact.Balance()
		log.Debugf(c, "Updated balance to %v | %v for contact #%d", balance, cpBalance, contact.ID)
	}

	if contactTransfersInfo := contact.GetTransfersInfo(); contactTransfersInfo.Last.ID != transfer.ID {
		contactTransfersInfo.Count += 1
		contactTransfersInfo.Last.ID = transfer.ID
		contactTransfersInfo.Last.At = transfer.DtCreated
		if transfer.HasInterest() {
			contactTransfersInfo.OutstandingWithInterest = append(contactTransfersInfo.OutstandingWithInterest, models.TransferWithInterestJson{
				TransferID:       transfer.ID,
				Amount:           transfer.AmountInCents,
				Currency:         transfer.Currency,
				Starts:           transfer.DtCreated,
				TransferInterest: transfer.TransferInterest,
			})
		}
		log.Debugf(c, "len(contactTransfersInfo.OutstandingWithInterest): %v", len(contactTransfersInfo.OutstandingWithInterest))
		if len(contactTransfersInfo.OutstandingWithInterest) > 0 {
			if len(closedTransferIDs) > 0 {
				log.Debugf(c, "removeClosedTransfersFromOutstandingWithInterest(closedTransferIDs: %v)", closedTransferIDs)
				contactTransfersInfo.OutstandingWithInterest = removeClosedTransfersFromOutstandingWithInterest(contactTransfersInfo.OutstandingWithInterest, closedTransferIDs)
			}
			log.Debugf(c, "transfer.ReturnToTransferIDs: %v", transfer.ReturnToTransferIDs)

			isClosed := func(transferID int64) bool {
				return slices.IsInInt64Slice(transferID, closedTransferIDs)
			}

		OuterLoop:
			for _, returnToTransferID := range transfer.ReturnToTransferIDs {
				if isClosed(returnToTransferID) {
					log.Debugf(c, "transfer %v is closed", returnToTransferID)
					continue
				}
				for i, outstanding := range contactTransfersInfo.OutstandingWithInterest {
					if outstanding.TransferID == returnToTransferID {
						if len(transfer.ReturnToTransferIDs) == 1 {
							outstanding.Returns = append(outstanding.Returns, models.TransferReturnJson{
								TransferID: transfer.ID,
								Amount:     transfer.AmountInCents,
								Time:       transfer.DtCreated,
							})
							contactTransfersInfo.OutstandingWithInterest[i] = outstanding
						} else {
							err = errors.WithMessage(ErrNotImplemented, "Return to multiple debts if at least one of them have interest is not implemented yet, please return debts with interest one by one.")
							return
						}
						continue OuterLoop
					}
				}
				log.Debugf(c, "transfer %v is not listed in contactTransfersInfo.OutstandingWithInterest", returnToTransferID)
			}
		}

		log.Debugf(c, "transfer.HasInterest(): %v, contactTransfersInfo: %v", transfer.HasInterest(), litter.Sdump(*contactTransfersInfo))
		if err = contact.SetTransfersInfo(*contactTransfersInfo); err != nil {
			err = errors.WithMessage(err, "failed to call SetTransfersInfo()")
			return
		}
	}
	return
}

func removeClosedTransfersFromOutstandingWithInterest(
	transfersWithInterest []models.TransferWithInterestJson,
	closedTransferIDs []int64,
) []models.TransferWithInterestJson {
	var i int
	for _, outstanding := range transfersWithInterest {
		if !slices.IsInInt64Slice(outstanding.TransferID, closedTransferIDs) {
			transfersWithInterest[i] = outstanding
			i += 1
		}
	}
	return transfersWithInterest[:i]
}

func InsertTransfer(c context.Context, transferEntity *models.TransferEntity) (transfer models.Transfer, err error) {
	transfer.TransferEntity = transferEntity
	err = dtdal.DB.InsertWithRandomIntID(c, &transfer)
	return
}

func (transferFacade) UpdateTransferOnReturn(c context.Context, returnTransfer, transfer models.Transfer, returnedAmount decimal.Decimal64p2) (err error) {
	log.Debugf(c, "UpdateTransferOnReturn(\n\treturnTransfer=%v,\n\ttransfer=%v,\n\treturnedAmount=%v)", litter.Sdump(returnTransfer), litter.Sdump(transfer), returnedAmount)

	if returnTransfer.Currency != transfer.Currency {
		panic(fmt.Sprintf("returnTransfer(id=%v).Currency != transfer.Currency => %v != %v", returnTransfer.ID, returnTransfer.Currency, transfer.Currency))
	} else if cID := returnTransfer.From().ContactID; cID != 0 && cID != transfer.To().ContactID {
		if transfer.To().ContactID == 0 && returnTransfer.From().UserID == transfer.To().UserID {
			transfer.To().ContactID = cID
			log.Warningf(c, "Fixed Transfer(%v).To().ContactID: 0 => %v", transfer.ID, cID)
		} else {
			panic(fmt.Sprintf("returnTransfer(id=%v).From().ContactID != transfer.To().ContactID => %v != %v", returnTransfer.ID, cID, transfer.To().ContactID))
		}
	} else if cID := returnTransfer.To().ContactID; cID != 0 && cID != transfer.From().ContactID {
		if transfer.From().ContactID == 0 && returnTransfer.To().UserID == transfer.From().UserID {
			transfer.From().ContactID = cID
			log.Warningf(c, "Fixed Transfer(%v).From().ContactID: 0 => %v", transfer.ID, cID)
		} else {
			panic(fmt.Sprintf("returnTransfer(id=%v).To().ContactID != transfer.From().ContactID => %v != %v", returnTransfer.ID, cID, transfer.From().ContactID))
		}
	}

	for _, previousReturn := range transfer.GetReturns() {
		if previousReturn.TransferID == returnTransfer.ID {
			log.Infof(c, "Transfer already has information about return transfer")
			return
		}
	}

	if outstandingValue := transfer.GetOutstandingValue(returnTransfer.DtCreated); outstandingValue < returnedAmount {
		log.Errorf(c, "transfer.GetOutstandingValue() < returnedAmount: %v <  %v", outstandingValue, returnedAmount)
		if outstandingValue <= 0 {
			return
		}
		returnedAmount = outstandingValue
	}

	if err = transfer.AddReturn(models.TransferReturnJson{
		TransferID: returnTransfer.ID,
		Time:       returnTransfer.DtCreated, // TODO: Replace with DtActual?
		Amount:     returnedAmount,
	}); err != nil {
		return
	}

	transfer.IsOutstanding = transfer.GetOutstandingValue(time.Now()) > 0

	if err = Transfers.SaveTransfer(c, transfer); err != nil {
		return
	}

	if dtdal.Reminder != nil {
		if err = dtdal.Reminder.DelayDiscardReminders(c, []int64{transfer.ID}, returnTransfer.ID); err != nil {
			err = errors.WithMessage(err, "failed to delay task to discard reminders")
			return
		}
	}

	return
}
