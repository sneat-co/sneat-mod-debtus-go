package facade

import (
	"fmt"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
)

type usersLinkingDbChanges struct {
	// use pointer as we pass it to FlagAsChanged() and IsChanged()
	db.Changes
	inviterUser, invitedUser       *models.AppUser
	inviterContact, invitedContact *models.Contact
}

func newUsersLinkingDbChanges() *usersLinkingDbChanges {
	return &usersLinkingDbChanges{}
}

type receiptDbChanges struct {
	// use pointer as we pass it to FlagAsChanged() and IsChanged()
	*usersLinkingDbChanges
	receipt  *models.Receipt
	transfer *models.Transfer
}

func newReceiptDbChanges() *receiptDbChanges {
	return &receiptDbChanges{
		usersLinkingDbChanges: newUsersLinkingDbChanges(),
	}
}

func workaroundReinsertContact(c context.Context, receipt models.Receipt, invitedContact models.Contact, changes *receiptDbChanges) (err error) {
	if _, err = GetContactByID(c, invitedContact.ID); err != nil {
		if db.IsNotFound(err) {
			log.Warningf(c, "workaroundReinsertContact(invitedContact.ID=%v) => %v", invitedContact.ID, err.Error())
			err = nil
			if receipt.Status == models.ReceiptStatusAcknowledged {
				if invitedContactInfo := changes.invitedUser.ContactByID(invitedContact.ID); invitedContactInfo != nil {
					log.Warningf(c, "Transactional retry, contact was not created in DB but invitedUser already has the contact info & receipt is acknowledged")
					changes.invitedContact = &invitedContact
				} else {
					log.Warningf(c, "Transactional retry, contact was not created in DB but receipt is acknowledged & invitedUser has not contact info in JSON")
				}
			}
			changes.FlagAsChanged(changes.invitedContact)
		} else {
			log.Errorf(c, "workaroundReinsertContact(invitedContact.ID=%v) => %v", invitedContact.ID, err.Error())
		}
	} else {
		log.Debugf(c, "workaroundReinsertContact(%v) => contact found by ID!", invitedContact.ID)
	}
	return
}

func AcknowledgeReceipt(
	c context.Context, receiptID, currentUserID int64, operation string,
) (
	receipt models.Receipt, transfer models.Transfer, isCounterpartiesJustConnected bool, err error,
) {
	log.Debugf(c, "AcknowledgeReceipt(receiptID=%d, currentUserID=%d, operation=%v)", receiptID, currentUserID, operation)
	var transferAckStatus string
	switch operation {
	case dal.AckAccept:
		transferAckStatus = models.TransferAccepted
	case dal.AckDecline:
		transferAckStatus = models.TransferDeclined
	default:
		err = ErrInvalidAcknowledgeType
		return
	}

	var invitedContact models.Contact

	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		var inviterUser, invitedUser models.AppUser
		var inviterContact models.Contact

		receipt, transfer, inviterUser, invitedUser, err = getReceiptTransferAndUsers(tc, receiptID, currentUserID)
		if err != nil {
			return
		}

		if transfer.CreatorUserID == currentUserID {
			log.Errorf(tc, "An attempt to claim receipt on self created transfer")
			err = ErrSelfAcknowledgement
			return
		}

		changes := &receiptDbChanges{
			receipt:  &receipt,
			transfer: &transfer,
			usersLinkingDbChanges: &usersLinkingDbChanges{
				inviterUser: &inviterUser,
				invitedUser: &invitedUser,
			},
		}

		if invitedContact.ID != 0 { // This means we are attempting to retry failed transaction
			if err = workaroundReinsertContact(tc, receipt, invitedContact, changes); err != nil {
				return
			}
		}

		{ // data integrity checks
			for _, counterpartyTgUserID := range invitedUser.GetTelegramUserIDs() {
				for _, creatorTgUserID := range inviterUser.GetTelegramUserIDs() {
					if counterpartyTgUserID == creatorTgUserID {
						return fmt.Errorf("data integrity issue: counterpartyTgUserID == creatorTgUserID (%v)", counterpartyTgUserID)
					}
				}
			}
		}

		if receipt.Status == models.ReceiptStatusAcknowledged {
			if receipt.AcknowledgedByUserID != currentUserID {
				err = fmt.Errorf("receipt.AcknowledgedByUserID != currentUserID (%d != %d)", receipt.AcknowledgedByUserID, currentUserID)
				return
			}
			log.Debugf(c, "Receipt is already acknowledged")
		} else {
			receipt.DtAcknowledged = time.Now()
			receipt.Status = models.ReceiptStatusAcknowledged
			receipt.AcknowledgedByUserID = currentUserID
			markReceiptAsViewed(receipt.ReceiptEntity, currentUserID)
			changes.FlagAsChanged(changes.receipt)

			transfer.AcknowledgeStatus = transferAckStatus
			transfer.AcknowledgeTime = receipt.DtAcknowledged
			changes.FlagAsChanged(changes.transfer)
		}

		if transfer.Counterparty().UserID == 0 {
			if isCounterpartiesJustConnected, err = NewReceiptUsersLinker(changes).linkUsersByReceiptWithinTransaction(c, tc); err != nil {
				return
			}
			invitedContact = *changes.invitedContact
			inviterContact = *changes.inviterContact
			log.Debugf(c, "linkUsersByReceiptWithinTransaction() =>\n\tinvitedContact %v: %+v\n\tinviterContact %v: %v",
				invitedContact.ID, invitedContact.ContactEntity, inviterContact.ID, inviterContact.ContactEntity)
		} else {
			log.Debugf(c, "No need to link users as already linked")
			inviterContact.ID = transfer.CounterpartyInfoByUserID(inviterUser.ID).ContactID
			invitedContact.ID = transfer.CounterpartyInfoByUserID(invitedUser.ID).ContactID
		}

		inviterUser.CountOfAckTransfersByCounterparties += 1
		invitedUser.CountOfAckTransfersByUser += 1

		if entitiesToSave := changes.EntityHolders(); len(entitiesToSave) > 0 {
			log.Debugf(c, "%v entities to save: %+v", len(entitiesToSave), entitiesToSave)
			if err = dal.DB.UpdateMulti(c, entitiesToSave); err != nil {
				return
			}
		} else {
			log.Debugf(c, "Nothing to save")
		}

		//if _, err = GetContactByID(c, invitedContact.ID); err != nil {
		//	if db.IsNotFound(err) {
		//		log.Errorf(c, "Invited contact is not found by ID, let's try to re-insert.")
		//		if err = facade.SaveContact(c, invitedContact); err != nil {
		//			return
		//		}
		//	} else {
		//		return
		//	}
		//}
		return
	}, dal.CrossGroupTransaction)

	if err != nil {
		if err == ErrSelfAcknowledgement {
			err = nil
			return
		}
		err = errors.WithMessage(err, "failed to acknowledge receipt")
		return
	}
	log.Infof(c, "Receipt successfully acknowledged")

	{ // verify invitedContact
		if invitedContact, err = GetContactByID(c, invitedContact.ID); err != nil {
			err = errors.WithMessage(err, "failed to load invited contact outside of transaction")
			if db.IsNotFound(err) {
				return
			}
			log.Errorf(c, err.Error())
			err = nil // We are OK to ignore technical issues here
			return
		}
	}
	return
}

func MarkReceiptAsViewed(c context.Context, receiptID, userID int64) (receipt models.Receipt, err error) {
	err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
		receipt, err = dal.Receipt.GetReceiptByID(tc, receiptID)
		if err != nil {
			return err
		}
		changed := markReceiptAsViewed(receipt.ReceiptEntity, userID)

		if receipt.DtViewed.IsZero() {
			receipt.DtViewed = time.Now()
			changed = true
		}
		if changed {
			dal.Receipt.UpdateReceipt(c, receipt)
		}
		return err
	}, dal.CrossGroupTransaction)
	return
}

func markReceiptAsViewed(receipt *models.ReceiptEntity, userID int64) (changed bool) {
	alreadyViewedByUser := false
	for _, uid := range receipt.ViewedByUserIDs {
		if uid == userID {
			alreadyViewedByUser = true
			break
		}
	}
	if !alreadyViewedByUser {
		receipt.ViewedByUserIDs = append(receipt.ViewedByUserIDs, userID)
		changed = true
	}
	return
}

func getReceiptTransferAndUsers(c context.Context, receiptID, userID int64) (
	receipt models.Receipt,
	transfer models.Transfer,
	creatorUser models.AppUser,
	counterpartyUser models.AppUser,
	err error,
) {
	log.Debugf(c, "getReceiptTransferAndUsers(receiptID=%v, userID=%v)", receiptID, userID)

	if receipt, err = dal.Receipt.GetReceiptByID(c, receiptID); err != nil {
		return
	}

	if transfer, err = GetTransferByID(c, receipt.TransferID); err != nil {
		return
	}

	if receipt.CreatorUserID != transfer.CreatorUserID {
		err = errors.New("Data integrity issue: receipt.CreatorUserID != transfer.CreatorUserID")
		return
	}

	if creatorUser, err = User.GetUserByID(c, transfer.CreatorUserID); err != nil {
		return
	}

	if counterpartyUser.ID = transfer.Counterparty().UserID; counterpartyUser.ID == 0 && userID != creatorUser.ID {
		counterpartyUser.ID = userID
	}

	if counterpartyUser.ID != 0 {
		if counterpartyUser, err = User.GetUserByID(c, counterpartyUser.ID); err != nil {
			return
		}
	}

	log.Debugf(c, "getReceiptTransferAndUsers(receiptID=%v, userID=%v) =>\n\tcreatorUser(id=%v): %+v\n\tcounterpartyUser(id=%v): %+v",
		receiptID, userID,
		creatorUser.ID, creatorUser.AppUserEntity,
		counterpartyUser.ID, counterpartyUser.AppUserEntity,
	)

	if creatorUser.AppUserEntity == nil {
		err = fmt.Errorf("creatorUser(id=%v) == nil - data integrity or app logic issue", transfer.CreatorUserID)
		return
	}
	return
}
