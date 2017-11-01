package gaedal

import (
	"golang.org/x/net/context"
	"github.com/strongo/decimal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/app/db"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"google.golang.org/appengine/delay"
	"github.com/strongo/app/gae"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/app/log"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
)

func (_ TransferDalGae) DelayUpdateTransfersOnReturn(c context.Context, returnTransferID int64, transferReturnsUpdate []dal.TransferReturnUpdate) (err error) {
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

	return dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if transfer, err = dal.Transfer.GetTransferByID(c, transferID); err != nil {
			if db.IsNotFound(err) {
				log.Errorf(c, err.Error())
				err = nil
			}
			return
		}
		return dal.Transfer.UpdateTransferOnReturn(c, returnTransfer, transfer, returnedAmount)
	}, db.SingleGroupTransaction)
}

func (_ TransferDalGae) UpdateTransferOnReturn(c context.Context, returnTransfer, transfer models.Transfer, returnedAmount decimal.Decimal64p2) (err error) {
	log.Debugf(c, "UpdateTransferOnReturn(\n\treturnTransfer=%v,\n\ttransfer.ID=%v,\n\treturnedAmount=%v)", litter.Sdump(returnTransfer), litter.Sdump(transfer), returnedAmount)

	if returnTransfer.Currency != transfer.Currency {
		panic(fmt.Sprintf("returnTransfer.Currency != transfer.Currency => %v != %v", returnTransfer.Currency, transfer.Currency))
	} else if cID := returnTransfer.From().ContactID; cID != 0 && cID != transfer.To().ContactID {
		if transfer.To().ContactID == 0 && returnTransfer.From().UserID == transfer.To().UserID {
			transfer.To().ContactID = cID
			log.Warningf(c, "Fixed Transfer(%v).To().ContactID: 0 => %v", transfer.ID, cID)
		} else {
			panic(fmt.Sprintf("returnTransfer.From().ContactID != transfer.To().ContactID => %v != %v", cID, transfer.To().ContactID))
		}
	} else if cID := returnTransfer.To().ContactID; cID != 0 && cID != transfer.From().ContactID {
		if transfer.From().ContactID == 0 && returnTransfer.To().UserID == transfer.From().UserID {
			transfer.From().ContactID = cID
			log.Warningf(c, "Fixed Transfer(%v).From().ContactID: 0 => %v", transfer.ID, cID)
		} else {
			panic(fmt.Sprintf("returnTransfer.To().ContactID != transfer.From().ContactID => %v != %v", cID, transfer.From().ContactID))
		}
	}

	for _, id := range transfer.ReturnTransferIDs {
		if id == returnTransfer.ID {
			log.Infof(c, "Transfer already has information about return transfer")
			return
		}
	}

	if transfer.AmountInCentsOutstanding < returnedAmount {
		log.Errorf(c, "transfer.AmountInCentsOutstanding < returnedAmount (%v <  %v0", transfer.AmountInCentsOutstanding, returnedAmount)
		if transfer.AmountInCentsOutstanding <= 0 {
			return
		}
		returnedAmount = transfer.AmountInCentsOutstanding
	}

	transfer.ReturnTransferIDs = append(transfer.ReturnTransferIDs, returnTransfer.ID)
	transfer.AmountInCentsOutstanding -= returnedAmount
	transfer.AmountInCentsReturned += returnedAmount
	if transfer.AmountInCentsOutstanding == 0 {
		transfer.IsOutstanding = true
	}

	if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
		return
	}

	if err = dal.Reminder.DelayDiscardReminders(c, []int64{transfer.ID}, returnTransfer.ID); err != nil {
		err = errors.WithMessage(err, "failed to delay task to discard reminders")
		return
	}

	return
}