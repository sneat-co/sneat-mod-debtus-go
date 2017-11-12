package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"time"
)

type ReceiptUsersLinker struct {
}

func (linker ReceiptUsersLinker) LinkReceiptUsers(c context.Context, receiptID, counterpartyUserID int64) error {
	return dal.DB.RunInTransaction(c, func(tc context.Context) error {
		receipt, transfer, creatorUser, counterpartyUser, err := getReceiptTransferAndUsers(tc, receiptID, counterpartyUserID)

		_, err = linker.linkUsersByReceiptWithinTransaction(c, tc, receipt, transfer, creatorUser, counterpartyUser)
		return err
	}, dal.CrossGroupTransaction)
}

func (linker ReceiptUsersLinker) linkUsersByReceiptWithinTransaction(
	c, tc context.Context, // 'tc' is transactional context, 'c' is not
	receipt models.Receipt,
	transfer models.Transfer,
	inviterUser, invitedUser models.AppUser,
) (
	isCounterpartiesJustConnected bool,
	err error,
) {
	log.Debugf(c,
		"ReceiptUsersLinker.linkUsersByReceiptWithinTransaction(receipt.ID=%d, transfer.ID=%d, inviterUser.ID=%d, invitedUser.ID=%d)",
		receipt.ID, transfer.ID, inviterUser.ID, invitedUser.ID)

	if !dal.DB.IsInTransaction(tc) {
		err = errors.New("linkUsersByReceiptWithinTransaction is called outside of transaction")
		return
	}

	if err = linker.validateInput(receipt, transfer, inviterUser, invitedUser); err != nil {
		return
	}

	log.Debugf(c, "transferEntity: %v", transfer.TransferEntity)
	log.Debugf(c, "transfer.From(): %v", transfer.From())
	log.Debugf(c, "transfer.To(): %v", transfer.To())
	fromOriginal := *transfer.From()
	toOriginal := *transfer.To()

	if transfer.Creator().UserID != inviterUser.ID {
		panic("transfer.Creator().UserID != inviterUser.ID - invalid logic?")
	}

	transferCreatorCounterparty := transfer.Counterparty()

	usersLinker := UsersLinker{}

	var (
		inviterContact, invitedContact models.Contact
	)
	inviterContact, err = dal.Contact.GetContactByID(tc, transferCreatorCounterparty.ContactID)
	if err != nil {
		err = errors.Wrapf(err, "Failed to call dal.Contact.GetContactByID(%d)", transfer.Counterparty().ContactID)
		return
	}

	var entitiesToSave []db.EntityHolder

	entitiesToSave, invitedContact, err = usersLinker.LinkUsersWithinTransaction(c, tc, inviterUser, invitedUser, inviterContact)
	if err != nil {
		err = errors.Wrapf(err, "Failed to link users")
		return
	}

	// Update entities
	{
		var receiptChanged, transferChanged bool

		if receiptChanged, err = linker.updateReceipt(receipt, invitedUser); err != nil {
			return
		} else if receiptChanged {
			entitiesToSave = append(entitiesToSave, &receipt)
		}
		if transferChanged, err = linker.updateTransfer(transfer, inviterUser, invitedUser, inviterContact, invitedContact); err != nil {
			return
		} else if transferChanged {
			entitiesToSave = append(entitiesToSave, &transfer)
			log.Debugf(c, "transfer.From(): %v", transfer.From())
			log.Debugf(c, "transfer.To(): %v", transfer.To())
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

	if len(entitiesToSave) > 0 {
		if err = dal.DB.UpdateMulti(tc, entitiesToSave); err != nil {
			return
		}
	}

	if transfer.DtDueOn.After(time.Now()) {
		if err := dal.Reminder.DelayCreateReminderForTransferUser(tc, receipt.TransferID, transfer.Counterparty().UserID); err != nil {
			return isCounterpartiesJustConnected, errors.Wrap(err, "Failed to delay creation of reminder for transfer coutnerparty")
		}
	} else {
		if transfer.DtDueOn.IsZero() {
			log.Debugf(tc, "No neeed to create reminder for counterparty as no due date")
		} else {
			log.Debugf(tc, "No neeed to create reminder for counterparty as due date in past")
		}
	}
	return isCounterpartiesJustConnected, err
}

func (linker ReceiptUsersLinker) validateInput(
	receipt models.Receipt,
	transfer models.Transfer,
	creatorUser, counterpartyUser models.AppUser,
) error {
	if receipt.CounterpartyUserID != 0 {
		if receipt.CounterpartyUserID != counterpartyUser.ID { // Already linked
			return errors.New("An attempt to link 3d user to a receipt")
		}

		transferCounterparty := transfer.Counterparty()

		if transferCounterparty.UserID != 0 && transferCounterparty.UserID != counterpartyUser.ID {
			return errors.New(
				fmt.Sprintf(
					"transferCounterparty.UserID != counterpartyUser.ID : %d != %d",
					transferCounterparty.UserID, counterpartyUser.ID,
				),
			)
		}
	}
	return nil
}

func (linker ReceiptUsersLinker) updateReceipt(receipt models.Receipt, counterpartyUser models.AppUser) (receiptChanged bool, err error) {
	if receipt.CounterpartyUserID != counterpartyUser.ID {
		receipt.CounterpartyUserID = counterpartyUser.ID
		receiptChanged = true
	}
	return
}

func (linker ReceiptUsersLinker) updateTransfer(
	transfer models.Transfer,
	inviterUser, invitedUser models.AppUser,
	inviterContact, invitedContact models.Contact,
) (
	transferChanged bool, err error,
) {
	// Validate input parameters
	{
		if transfer.ID == 0 || transfer.TransferEntity == nil {
			panic(fmt.Sprintf("Invalid parameter: transfer: %v", transfer))
		}
		if inviterUser.ID == 0 || inviterUser.AppUserEntity == nil {
			panic(fmt.Sprintf("Invalid parameter: inviterUser: %v", inviterUser))
		}
		if invitedUser.ID == 0 || invitedUser.AppUserEntity == nil {
			panic(fmt.Sprintf("Invalid parameter: invitedUser: %v", invitedUser))
		}
		if inviterContact.ID == 0 || inviterContact.ContactEntity == nil {
			panic(fmt.Sprintf("Invalid parameter: inviterContact: %v", inviterContact))
		}
		if invitedContact.ID == 0 || invitedContact.ContactEntity == nil {
			panic(fmt.Sprintf("Invalid parameter: invitedContact: %v", invitedContact))
		}
	}

	transferCounterparty := transfer.Counterparty()

	if transferCounterparty.UserID != invitedUser.ID {
		if transferCounterparty.UserID != 0 {
			err = errors.New(fmt.Sprintf("transfer.Contact().UserID != counterpartyUserID : %d != %d",
				transfer.Counterparty().UserID, invitedUser.ID))
			return
		}
		transfer.Counterparty().UserID = invitedUser.ID
		transferChanged = true
	}

	updateTransferCounterparty := func(
		counterparty *models.TransferCounterpartyInfo,
		user models.AppUser,
		contact models.Contact,
	) {
		if counterparty.UserID == 0 {
			counterparty.UserID = user.ID
		} else if counterparty.UserID != user.ID {
			panic(fmt.Sprintf("counterparty.UserID != user.ID : %d != %d", counterparty.UserID, user.ID))
		}
		counterparty.UserName = user.FullName()

		if counterparty.ContactID == 0 {
			counterparty.ContactID = contact.ID
		} else if counterparty.ContactID != contact.ID {
			panic(fmt.Sprintf("counterparty.ContactID != contact.ID : %d != %d", counterparty.UserID, user.ID))
		}
		counterparty.ContactName = contact.FullName()
	}

	from := transfer.From()
	to := transfer.To()
	switch transfer.Direction() {
	case models.TransferDirectionUser2Counterparty:
		updateTransferCounterparty(from, inviterUser, invitedContact)
		updateTransferCounterparty(to, invitedUser, inviterContact)
	case models.TransferDirectionCounterparty2User:
		updateTransferCounterparty(from, invitedUser, inviterContact)
		updateTransferCounterparty(to, inviterUser, invitedContact)
	default:
		err = errors.New(fmt.Sprintf("Unknown transfer.Direction(): %v", transfer.Direction()))
		return
	}
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
