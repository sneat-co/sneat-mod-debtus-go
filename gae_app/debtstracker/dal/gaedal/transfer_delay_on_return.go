package gaedal

import (
	"fmt"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
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

	if returnTransfer, err = facade.GetTransferByID(c, returnTransferID); err != nil {
		if db.IsNotFound(err) {
			log.Errorf(c, errors.WithMessage(err, "return transfer not found").Error())
			err = nil
		}
		return
	}

	if transfer, err = facade.GetTransferByID(c, transferID); err != nil {
		return
	}
	var txOptions db.RunOptions
	if transfer.HasInterest() {
		txOptions = db.CrossGroupTransaction
	} else {
		txOptions = db.SingleGroupTransaction
	}

	return dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if transfer, err = facade.GetTransferByID(c, transferID); err != nil {
			if db.IsNotFound(err) {
				log.Errorf(c, err.Error())
				err = nil
			}
			return
		}
		if err = facade.Transfers.UpdateTransferOnReturn(c, returnTransfer, transfer, returnedAmount); err != nil {
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
			if user, err = facade.User.GetUserByID(c, userID); err != nil {
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
						err = facade.User.SaveUser(c, user)
					}
				}
			}
			return
		}
		removeFromContact := func() (err error) {
			var (
				contact models.Contact
			)
			if contact, err = facade.GetContactByID(c, contactID); err != nil {
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
					return facade.SaveContact(c, contact)
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
