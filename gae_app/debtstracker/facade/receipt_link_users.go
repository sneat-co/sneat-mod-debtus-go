package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"time"
	"github.com/strongo/app/slices"
)

type receiptUsersLinker struct {
	changes *receiptDbChanges
}

func NewReceiptUsersLinker(changes *receiptDbChanges) receiptUsersLinker {
	if changes == nil {
		changes = newReceiptDbChanges()
	}
	return receiptUsersLinker{
		changes: changes,
	}
}

func (linker *receiptUsersLinker) LinkReceiptUsers(c context.Context, receiptID, counterpartyUserID int64) (isJustLinked bool, err error) {
	log.Debugf(c, "receiptUsersLinker.LinkReceiptUsers(receiptID=%v, counterpartyUserID=%v)", receiptID, counterpartyUserID)
	var invitedContact models.Contact
	attempt := 0
	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		if attempt += 1; attempt > 1 {
			sleepPeriod := time.Duration(attempt) * time.Second
			log.Warningf(c, "Transaction retry will sleep for %v, invitedContact.ID: %v", attempt, invitedContact.ID)
			time.Sleep(sleepPeriod)
		}
		changes := linker.changes
		var (
			receipt models.Receipt
			transfer models.Transfer
			creatorUser models.AppUser
			counterpartyUser models.AppUser
		)
		if receipt, transfer, creatorUser, counterpartyUser, err = getReceiptTransferAndUsers(tc, receiptID, counterpartyUserID); err != nil {
			return
		}
		changes.receipt = &receipt
		changes.transfer = &transfer
		changes.inviterUser = &creatorUser
		changes.invitedUser = &counterpartyUser
		if invitedContact.ID != 0 { // This means we are attempting to retry failed transaction
			if err = workaroundReinsertContact(tc, receipt, invitedContact, changes); err != nil {
				return
			}
		}

		if isJustLinked, err = linker.linkUsersByReceiptWithinTransaction(c, tc); err != nil {
			return
		} else {
			invitedContact = *changes.invitedContact
		}

		// Integrity checks
		{
			invitedContact.MustMatchCounterparty(*changes.inviterContact)
		}

		if entitiesToSave := changes.EntityHolders(); len(entitiesToSave) > 0 {
			if err = dal.DB.UpdateMulti(c, entitiesToSave); err != nil {
				return
			}
		} else {
			log.Debugf(c, "Receipt and transfer has not changed")
		}
		return
	}, dal.CrossGroupTransaction)
	if err != nil {
		return
	}
	log.Debugf(c, "receiptUsersLinker.LinkReceiptUsers() => invitedContact: %+v", invitedContact)
	if invitedContact, err = dal.Contact.GetContactByID(c, invitedContact.ID); err != nil {
		return
	}
	log.Debugf(c, "receiptUsersLinker.LinkReceiptUsers() => invitedContact from DB: %+v", invitedContact)
	return
}

func (linker receiptUsersLinker) linkUsersByReceiptWithinTransaction(
	c, tc context.Context, // 'tc' is transactional context, 'c' is not
) (
	isCounterpartiesJustConnected bool,
	err error,
) {
	if !dal.DB.IsInTransaction(tc) {
		panic("linkUsersByReceiptWithinTransaction is called outside of transaction")
	}

	changes := linker.changes
	receipt := changes.receipt
	transfer := changes.transfer
	inviterUser, invitedUser := *changes.inviterUser, *changes.invitedUser
	var invitedContact, inviterContact models.Contact
	if changes.inviterContact != nil {
		inviterContact = *changes.inviterContact
	}
	if changes.invitedContact != nil {
		invitedContact = *changes.invitedContact
	}

	log.Debugf(c,
		"receiptUsersLinker.linkUsersByReceiptWithinTransaction(receipt.ID=%d, transfer.ID=%d, inviterUser.ID=%d, invitedUser.ID=%d, inviterContact.ID=%v, invitedContact.ID=%v)",
		receipt.ID, transfer.ID, inviterUser.ID, invitedUser.ID, inviterContact.ID, invitedContact.ID)

	{ // validate inputs
		if err = linker.validateInput(changes); err != nil {
			return
		}
		if receipt.TransferID != transfer.ID {
			panic(fmt.Sprintf("receipt.TransferID != transfer.ID: %v != %v", receipt.TransferID, transfer.ID))
		}
		if transferCreatorUserID := transfer.Creator().UserID; transferCreatorUserID == 0 {
			panic("transfer.Creator().UserID is zero")
		} else if transferCreatorUserID != inviterUser.ID {
			panic(fmt.Sprintf("transfer.Creator().UserID != inviterUser.ID:  %v != %v", transferCreatorUserID, inviterUser.ID))
		} else if transferCreatorUserID == invitedUser.ID {
			panic(fmt.Sprintf("transfer.Creator().UserID == invitedUser.ID:  %v != %v", transferCreatorUserID, invitedUser.ID))
		}
	}

	log.Debugf(c, "transferEntity: %v", transfer.TransferEntity)
	log.Debugf(c, "transfer.From(): %v", transfer.From())
	log.Debugf(c, "transfer.To(): %v", transfer.To())
	fromOriginal := *transfer.From()
	toOriginal := *transfer.To()

	transferCreatorCounterparty := transfer.Counterparty()

	if inviterContact, err = dal.Contact.GetContactByID(tc, transferCreatorCounterparty.ContactID); err != nil {
		return
	} else if inviterContact.UserID !=  inviterUser.ID {
		panic(fmt.Sprintf("inviterContact.UserID !=  inviterUser.ID: %v != %v", inviterContact.UserID, inviterUser.ID))
	} else {
		changes.inviterContact = &inviterContact
	}

	if err = newUsersLinker(changes.usersLinkingDbChanges).linkUsersWithinTransaction(tc); err != nil {
		err = errors.WithMessage(err, "Failed to link users")
		return
	} else {
		invitedContact = *changes.invitedContact // as was updated
	}
	{ // Update invited user's last currency
		var invitedUserChanged bool
		if invitedUser.LastCurrencies, invitedUserChanged = slices.MergeStrings(invitedUser.LastCurrencies, []string{string(transfer.Currency)}); invitedUserChanged {
			changes.FlagAsChanged(changes.invitedUser)
		}
	}

	log.Debugf(c, "linkUsersWithinTransaction() => invitedContact.ID: %v, inviterContact.ID: %v", invitedContact.ID, inviterContact.ID)

	// Update entities
	{
		if err = linker.updateReceipt(); err != nil {
			return
		} else
		if err = linker.updateTransfer(); err != nil {
			return
		} else if linker.changes.IsChanged(linker.changes.transfer) {
			log.Debugf(c, "transfer changed:\n\tFrom(): %v\n\tTo(): %v", transfer.From(), transfer.To())
			// Just double check we did not screw up
			{
				if fromOriginal.UserID != 0 && fromOriginal.UserID != transfer.From().UserID {
					err = errors.New("fromOriginal.UserID != 0 && fromOriginal.UserID != transfer.From().UserID")
					return
				}
				if fromOriginal.ContactID != 0 && fromOriginal.ContactID != transfer.From().ContactID {
					err = errors.New("fromOriginal.ContactID != 0 && fromOriginal.ContactID != transfer.From().ContactID")
					return
				}
				if toOriginal.UserID != 0 && toOriginal.UserID != transfer.To().UserID {
					err = errors.New("toOriginal.UserID != 0 && toOriginal.UserID != transfer.To().UserID")
					return
				}
				if toOriginal.ContactID != 0 && toOriginal.ContactID != transfer.To().ContactID {
					err = errors.New("toOriginal.ContactID != 0 && toOriginal.ContactID != transfer.To().ContactID")
					return
				}
			}
		}
	}

	if transfer.DtDueOn.After(time.Now()) {
		if err = dal.Reminder.DelayCreateReminderForTransferUser(tc, receipt.TransferID, transfer.Counterparty().UserID); err != nil {
			err = errors.WithMessage(err, "Failed to delay creation of reminder for transfer coutnerparty")
			return
		}
	} else {
		if transfer.DtDueOn.IsZero() {
			log.Debugf(tc, "No need to create reminder for counterparty as no due date")
		} else {
			log.Debugf(tc, "No need to create reminder for counterparty as due date in past")
		}
	}
	return
}

func (linker receiptUsersLinker) validateInput(changes *receiptDbChanges) error {

	if changes.receipt.CounterpartyUserID != 0 {
		if changes.receipt.CounterpartyUserID != changes.invitedUser.ID { // Already linked
			return errors.New("An attempt to link 3d user to a receipt")
		}

		transferCounterparty := changes.transfer.Counterparty()

		if transferCounterparty.UserID != 0 && transferCounterparty.UserID != changes.invitedUser.ID {
			return errors.New(
				fmt.Sprintf(
					"transferCounterparty.UserID != invitedUser.ID : %d != %d",
					transferCounterparty.UserID, changes.invitedUser.ID,
				),
			)
		}
	}
	return nil
}

func (linker receiptUsersLinker) updateReceipt() (err error) {
	receipt := *linker.changes.receipt
	counterpartyUser := *linker.changes.invitedUser
	if receipt.CounterpartyUserID != counterpartyUser.ID {
		receipt.CounterpartyUserID = counterpartyUser.ID
		linker.changes.FlagAsChanged(linker.changes.receipt)
	}
	return
}

func (linker receiptUsersLinker) updateTransfer() (err error) {
	changes := linker.changes

	transfer := changes.transfer
	inviterUser, invitedUser := *changes.inviterUser, *changes.invitedUser
	inviterContact, invitedContact := *changes.inviterContact, *changes.invitedContact
	{ // Validate input parameters
		if transfer.ID == 0 || transfer.TransferEntity == nil {
			panic(fmt.Sprintf("Invalid parameter: transfer: %v", transfer))
		}
		validateSide := func(side string, user models.AppUser, contact models.Contact) {
			if user.ID == 0 || user.AppUserEntity == nil {
				panic(fmt.Sprintf("receiptUsersLinker.updateTransfer() => %vUser: %v", side, user))
			}
			if contact.ID == 0 || contact.ContactEntity == nil {
				panic(fmt.Sprintf("receiptUsersLinker.updateTransfer() => %vContact: %v", side, contact))
			} else if contact.UserID != user.ID {
				panic(fmt.Sprintf("receiptUsersLinker.updateTransfer() => %vContact.UserID != %vUser.ID: %v != %v", side, side, contact.UserID, invitedUser.ID))
			}
		}
		validateSide("inviter", inviterUser, inviterContact)
		validateSide("invited", invitedUser, invitedContact)
		if transfer.CreatorUserID != inviterUser.ID {
			panic(fmt.Sprintf("receiptUsersLinker.updateTransfer() => transfer.CreatorUserID != inviterUser.ID: %v != %v", transfer.CreatorUserID, invitedUser.ID))
		}
	}

	transferCounterparty := transfer.Counterparty()

	if transferCounterparty.UserID != invitedUser.ID {
		if transferCounterparty.UserID != 0 {
			err = fmt.Errorf("transfer.Contact().UserID != counterpartyUserID : %d != %d",
				transfer.Counterparty().UserID, invitedUser.ID)
			return
		}
		transfer.Counterparty().UserID = invitedUser.ID
		linker.changes.FlagAsChanged(linker.changes.transfer)
	}

	updateTransferCounterpartyInfo := func(
		side string,
		counterparty *models.TransferCounterpartyInfo,
		user models.AppUser,
		contact models.Contact,
	) {
		if contact.UserID == user.ID {
			panic(fmt.Sprintf(
				"updateTransferCounterpartyInfo() => %vContact.UserID == %vUser.ID : %d, counterparty.UserID: %v",
				side, side, contact.UserID, counterparty.UserID))
		}
		if counterparty.UserID == 0 {
			counterparty.UserID = user.ID
		} else if counterparty.UserID != user.ID {
			panic(fmt.Sprintf("updateTransferCounterpartyInfo() => counterparty.UserID != %vUser.ID : %d != %d, %vContact.UserID: %v", side, counterparty.UserID, user.ID, side, contact.UserID))
		}
		counterparty.UserName = user.FullName()

		if counterparty.ContactID == 0 {
			counterparty.ContactID = contact.ID
		} else if counterparty.ContactID != contact.ID {
			panic(fmt.Sprintf(
				"receiptUsersLinker.updateTransfer() => counterparty.ContactID != %vContact.ID : %d != %d",
				side, counterparty.ContactID, contact.ID))
		}
		counterparty.ContactName = contact.FullName()
	}

	updateTransferCounterpartyInfo("inviter", transfer.Creator(), inviterUser, invitedContact)
	updateTransferCounterpartyInfo("invited", transfer.Counterparty(), invitedUser, inviterContact)

	//if inlineMessageID != "" {
	//	transfer.CounterpartyTgReceiptInlineMessageID = inlineMessageID
	//}
	transferAmount := transfer.GetAmount()
	transferVal := transferAmount.Value
	if transfer.Direction() == models.TransferDirectionUser2Counterparty {
		transferVal *= -1
	}

	return
}
