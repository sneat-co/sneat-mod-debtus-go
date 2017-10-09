package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"time"
	"github.com/strongo/app/db"
)

func AcknowledgeReceipt(c context.Context, receiptID, currentUserID int64, operation string) (receipt models.Receipt, transfer models.Transfer, isCounterpartiesJustConnected bool, err error) {
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

	var creatorUser, counterpartyUser models.AppUser

	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		receipt, transfer, creatorUser, counterpartyUser, err = getReceiptTransferAndUsers(tc, receiptID, currentUserID)
		if err != nil {
			return
		}

		if transfer.CreatorUserID == currentUserID {
			log.Errorf(tc, "An attempt to claim receipt on self created transfer")
			err = ErrSelfAcknowledgement
			return
		}

		if receipt.Status == models.ReceiptStatusAcknowledged {
			if receipt.AcknowledgedByUserID != currentUserID {
				err = errors.New(fmt.Sprintf("receipt.AcknowledgedByUserID != currentUserID (%d != %d)", receipt.AcknowledgedByUserID, currentUserID))
				return
			}
			return
		}

		for _, counterpartyUserID := range counterpartyUser.GetTelegramUserIDs() {
			for _, creatorUserID := range creatorUser.GetTelegramUserIDs() {
				if counterpartyUserID == creatorUserID {
					return errors.New(fmt.Sprintf("Data integrity issue: counterpartyUserID == creatorUserID (%v)", counterpartyUserID))
				}
			}
		}

		receipt.DtAcknowledged = time.Now()
		receipt.Status = models.ReceiptStatusAcknowledged
		receipt.AcknowledgedByUserID = currentUserID
		markReceiptAsViewed(receipt.ReceiptEntity, currentUserID)

		transfer.AcknowledgeStatus = transferAckStatus
		transfer.AcknowledgeTime = receipt.DtAcknowledged

		if transfer.Counterparty().UserID == 0 {
			isCounterpartiesJustConnected, err = ReceiptUsersLinker{}.linkUsersByReceiptWithinTransaction(c, tc, receipt, transfer, creatorUser, counterpartyUser)
			if err != nil {
				return
			}
		}

		creatorUser.CountOfAckTransfersByCounterparties += 1
		counterpartyUser.CountOfAckTransfersByUser += 1

		return dal.DB.UpdateMulti(c, []db.EntityHolder{&receipt, &transfer, &creatorUser, &counterpartyUser})
	}, dal.CrossGroupTransaction)

	if err != nil {
		err = errors.Wrap(err, "Failed to acknowledge receipt")
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

func getReceiptTransferAndUsers(c context.Context, receiptID, userID int64) (receipt models.Receipt, transfer models.Transfer, creatorUser, counterpartyUser models.AppUser, err error) {

	errCapacity := 4 // For getting currentUser, receipt, transfer entities in parallel.
	errs := make(chan error, errCapacity)

	var user, anotherUser models.AppUser

	go func() { // We load current user anyway - it can be a creator, counterparty, or non-authorized person
		var err error
		user, err = dal.User.GetUserByID(c, userID)
		errs <- err
	}()
	go func() {
		var (
			err error
		)
		receipt, err = dal.Receipt.GetReceiptByID(c, receiptID)
		if errs <- err; err != nil {
			return
		}

		transfer, err = dal.Transfer.GetTransferByID(c, receipt.TransferID)
		if errs <- err; err != nil {
			return
		}

		if receipt.CreatorUserID != transfer.CreatorUserID {
			errs <- errors.New("Data integrity issue: receipt.CreatorUserID != transfer.CreatorUserID")
			return
		}

		if userID == transfer.CreatorUserID {
			// If current user is creator of transfer the counterparty user can be loaded just if we know ID
			if transfer.Counterparty().UserID != 0 {
				anotherUser, err = dal.User.GetUserByID(c, transfer.Counterparty().UserID)
			}
		} else { // If current user is not creator
			if transfer.Counterparty().UserID != 0 && transfer.Counterparty().UserID != userID {
				err = errors.New(fmt.Sprintf("Attempt to access receipt(id=%v) & transfer(id=%v) by non related user(id=%v)\n\tCreatorUserID: %v, CounterpartUserID: %v",
					receiptID, receipt.TransferID, userID, transfer.CreatorUserID, transfer.Counterparty().UserID))
			} else {
				anotherUser, err = dal.User.GetUserByID(c, transfer.CreatorUserID)
			}
		}
		if errs <- err; err != nil {
			return
		}
	}()
	for i := 0; i < errCapacity; i++ {
		if err = <-errs; err != nil {
			return
		}
	}
	if transfer.CreatorUserID == userID {
		log.Debugf(c, "transfer.CreatorUserID == userID")
		creatorUser = user
		counterpartyUser = anotherUser
	} else {
		log.Debugf(c, "transfer.CreatorUserID != userID")
		creatorUser = anotherUser
		counterpartyUser = user
	}

	log.Debugf(c, "getReceiptTransferAndUsers(receiptID=%v, userID=%v):\n\tcreatorUser(%v): %v\n\tcounterpartyUser(%v): %v",
		receiptID, userID,
		receipt.CreatorUserID, creatorUser,
		transfer.Counterparty().UserID, counterpartyUser,
	)

	if creatorUser.AppUserEntity == nil {
		err = errors.New(fmt.Sprintf("creatorUser(id=%v) == nil - data integrity or app logic issue", transfer.CreatorUserID))
		return
	}
	return
}
