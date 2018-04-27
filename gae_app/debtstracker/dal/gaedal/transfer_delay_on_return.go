package gaedal

import (
	"fmt"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
)

func (TransferDalGae) DelayUpdateTransfersOnReturn(c context.Context, returnTransferID int64, transferReturnsUpdate []dal.TransferReturnUpdate) (err error) {
	log.Debugf(c, "DelayUpdateTransfersOnReturn(returnTransferID=%v, transferReturnsUpdate=%v)", returnTransferID, transferReturnsUpdate)
	if returnTransferID == 0 {
		panic("returnTransferID == 0")
	}
	if len(transferReturnsUpdate) == 0 {
		panic("len(transferReturnsUpdate) == 0")
	}
	for i, transferReturnUpdate := range transferReturnsUpdate {
		if transferReturnUpdate.TransferID == 0 {
			panic(fmt.Sprintf("transferReturnsUpdates[%d].TransferID == 0", i))
		}
		if transferReturnUpdate.ReturnedAmount <= 0 {
			panic(fmt.Sprintf("transferReturnsUpdates[%d].Amount <= 0: %v", i, transferReturnUpdate.ReturnedAmount))
		}
	}
	return gae.CallDelayFunc(c, common.QUEUE_TRANSFERS, "update-transfers-on-return", delayUpdateTransfersOnReturn, returnTransferID, transferReturnsUpdate)
}

var delayUpdateTransfersOnReturn = delay.Func("updateTransfersOnReturn", updateTransfersOnReturn)

func updateTransfersOnReturn(c context.Context, returnTransferID int64, transferReturnsUpdate []dal.TransferReturnUpdate) (err error) {
	log.Debugf(c, "updateTransfersOnReturn(returnTransferID=%v, transferReturnsUpdate=%+v)", returnTransferID, transferReturnsUpdate)
	for i, transferReturnUpdate := range transferReturnsUpdate {
		if transferReturnUpdate.TransferID == 0 {
			panic(fmt.Sprintf("transferReturnsUpdates[%d].TransferID == 0", i))
		}
		if transferReturnUpdate.ReturnedAmount <= 0 {
			panic(fmt.Sprintf("transferReturnsUpdates[%d].Amount <= 0: %v", i, transferReturnUpdate.ReturnedAmount))
		}
		if err = DelayUpdateTransferOnReturn(c, returnTransferID, transferReturnUpdate.TransferID, transferReturnUpdate.ReturnedAmount); err != nil {
			return
		}
	}
	return
}

func DelayUpdateTransferOnReturn(c context.Context, returnTransferID, transferID int64, returnedAmount decimal.Decimal64p2) error {
	return gae.CallDelayFunc(c, common.QUEUE_TRANSFERS, "update-transfer-on-return", delayUpdateTransferOnReturn, returnTransferID, transferID, returnedAmount)
}

var delayUpdateTransferOnReturn = delay.Func("updateTransferOnReturn", updateTransferOnReturn)

func updateTransferOnReturn(c context.Context, returnTransferID, transferID int64, returnedAmount decimal.Decimal64p2) (err error) {
	log.Debugf(c, "updateTransferOnReturn(returnTransferID=%v, transferID=%v, returnedAmount=%v)", returnTransferID, transferID, returnedAmount)

	var transfer, returnTransfer models.Transfer

	if returnTransfer, err = dal.Transfer.GetTransferByID(c, returnTransferID); err != nil {
		if db.IsNotFound(err) {
			log.Errorf(c, errors.WithMessage(err, "return transfer not found").Error())
			err = nil
		}
		return
	}

	if transfer, err = dal.Transfer.GetTransferByID(c, transferID); err != nil {
		return
	}
	var txOptions db.RunOptions
	if transfer.HasInterest() {
		txOptions = db.CrossGroupTransaction
	} else {
		txOptions = db.SingleGroupTransaction
	}

	return dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if transfer, err = dal.Transfer.GetTransferByID(c, transferID); err != nil {
			if db.IsNotFound(err) {
				log.Errorf(c, err.Error())
				err = nil
			}
			return
		}
		if err = dal.Transfer.UpdateTransferOnReturn(c, returnTransfer, transfer, returnedAmount); err != nil {
			return
		}
		if transfer.HasInterest() && !transfer.IsOutstanding {
			if err = removeFromOutstandingWithInterest(c, transfer); err != nil {
				return
			}
		}
		return
	}, txOptions)
}

func removeFromOutstandingWithInterest(c context.Context, transfer models.Transfer) (err error) {
	removeFromOutstanding := func(userID, contactID int64) (err error) {
		if userID == 0 && contactID == 0 {
			return
		} else if userID == 0 {
			panic("removeFromOutstandingWithInterest(): userID == 0")
		} else if contactID == 0 {
			panic("removeFromOutstandingWithInterest(): contactID == 0")
		}
		removeFromUser := func() (err error) {
			var (
				user models.AppUser
				//contact models.Contact
			)
			if user, err = dal.User.GetUserByID(c, userID); err != nil {
				return
			}
			contacts := user.Contacts()
			for _, userContact := range contacts {
				for i, outstanding := range userContact.Transfers.OutstandingWithInterest {
					if outstanding.TransferID == transfer.ID {
						// https://github.com/golang/go/wiki/SliceTricks
						a := userContact.Transfers.OutstandingWithInterest
						userContact.Transfers.OutstandingWithInterest = append(a[:i], a[i+1:]...)
						user.SetContacts(contacts)
						user.TransfersWithInterestCount -= 1
						err = dal.User.SaveUser(c, user)
					}
				}
			}
			return
		}
		removeFromContact := func() (err error) {
			var (
				contact models.Contact
			)
			if contact, err = dal.Contact.GetContactByID(c, contactID); err != nil {
				return
			}
			if contact.UserID != userID {
				return fmt.Errorf("contact.UserID != userID: %v != %v", contact.UserID, userID)
			}
			transfersInfo := *contact.GetTransfersInfo()
			for i, outstanding := range transfersInfo.OutstandingWithInterest {
				if outstanding.TransferID == transfer.ID {
					// https://github.com/golang/go/wiki/SliceTricks
					a := transfersInfo.OutstandingWithInterest
					transfersInfo.OutstandingWithInterest = append(a[:i], a[i+1:]...)
					if err = contact.SetTransfersInfo(transfersInfo); err != nil {
						return
					}
					return dal.Contact.SaveContact(c, contact)
				}
			}
			return
		}
		if err = removeFromUser(); err != nil {
			return
		}
		if err = removeFromContact(); err != nil {
			return
		}
		return
	}
	from, to := transfer.From(), transfer.To()

	if err = removeFromOutstanding(from.UserID, to.ContactID); err != nil {
		return
	}
	if err = removeFromOutstanding(to.UserID, from.ContactID); err != nil {
		return
	}
	return
}

func (TransferDalGae) UpdateTransferOnReturn(c context.Context, returnTransfer, transfer models.Transfer, returnedAmount decimal.Decimal64p2) (err error) {
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

	//transfer.ReturnTransferIDs = append(transfer.ReturnTransferIDs, returnTransfer.ID)
	returns := transfer.GetReturns()
	if len(returns) == 0 && len(transfer.ReturnTransferIDs) != 0 { // TODO: Remove fix for old transfers
		var returnTransfers []models.Transfer
		if returnTransfers, err = dal.Transfer.GetTransfersByID(c, transfer.ReturnTransferIDs); err != nil {
			return
		}
		returns = make([]models.TransferReturnJson, len(transfer.ReturnTransferIDs), len(transfer.ReturnTransferIDs)+1)
		var returnedVal decimal.Decimal64p2

		for i, rt := range returnTransfers {
			returns[i] = models.TransferReturnJson{
				TransferID: rt.ID,
				Time:       rt.DtCreated, // TODO: Replace with DtActual?
				Amount:     rt.AmountInCents,
			}
			returnedVal += rt.AmountInCents
		}
		if returnedVal > transfer.AmountInCents {
			log.Warningf(c, "failed to properly migrated ReturnTransferIDs to ReturnsJson: returnedAmount > transfer.AmountInCents")
			for i := range returns {
				returns[i].Amount = 0
			}
		}
	}
	returns = append(returns, models.TransferReturnJson{
		TransferID: returnTransfer.ID,
		Time:       returnTransfer.DtCreated, // TODO: Replace with DtActual?
		Amount:     returnedAmount,
	})
	transfer.SetReturns(returns)

	transfer.IsOutstanding = transfer.GetOutstandingValue(time.Now()) > 0

	if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
		return
	}

	if err = dal.Reminder.DelayDiscardReminders(c, []int64{transfer.ID}, returnTransfer.ID); err != nil {
		err = errors.WithMessage(err, "failed to delay task to discard reminders")
		return
	}

	return
}
