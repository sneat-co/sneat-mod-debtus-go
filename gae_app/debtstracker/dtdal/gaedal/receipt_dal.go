package gaedal

import (
	"errors"
	"github.com/dal-go/dalgo/dal"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/app/gae"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
	"google.golang.org/appengine/v2/delay"
)

func NewReceiptIncompleteKey() *dal.Key {
	return models.NewReceiptKey(0)
}

type ReceiptDalGae struct {
}

func NewReceiptDalGae() ReceiptDalGae {
	return ReceiptDalGae{}
}

var _ dtdal.ReceiptDal = (*ReceiptDalGae)(nil)

func (ReceiptDalGae) UpdateReceipt(c context.Context, tx dal.ReadwriteTransaction, receipt models.Receipt) error {
	return tx.Set(c, receipt.Record)
}

func (receiptDalGae ReceiptDalGae) GetReceiptByID(c context.Context, tx dal.ReadSession, id int) (receipt models.Receipt, err error) {
	receipt = models.NewReceipt(id, nil)
	return receipt, tx.Get(c, receipt.Record)
}

func (receiptDalGae ReceiptDalGae) CreateReceipt(c context.Context, receipt *models.ReceiptEntity) (id int, err error) { // TODO: Move to facade
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		receiptKey := NewReceiptIncompleteKey()
		user, err := facade.User.GetUserByID(c, receipt.CreatorUserID)
		if err != nil {
			return err
		}
		user.Data.CountOfReceiptsCreated += 1
		if keys, err := gaedb.PutMulti(c, []*datastore.Key{receiptKey, models.NewAppUserKey(receipt.CreatorUserID)}, []interface{}{receipt, user}); err != nil {
			return err
		} else {
			receiptKey = keys[0]
		}
		id = receiptKey.IntID()
		return nil
	})
	return
}

func (receiptDalGae ReceiptDalGae) MarkReceiptAsSent(c context.Context, receiptID, transferID int, sentTime time.Time) error {
	return errors.New("TODO: Implement MarkReceiptAsSent")
	//return dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
	//	var (
	//		receipt     models.Receipt
	//		transfer    models.Transfer
	//		transferKey *datastore.Key
	//	)
	//	receiptKey := NewReceiptKey(c, receiptID)
	//	if transferID == 0 {
	//		if receipt, err = receiptDalGae.GetReceiptByID(c, receiptID); err != nil {
	//			return err
	//		}
	//		if transfer, err = facade.Transfers.GetTransferByID(c, transferID); err != nil {
	//			return err
	//		}
	//		transferKey = NewTransferKey(c, transferID)
	//	} else {
	//		receipt.ReceiptEntity = new(models.ReceiptEntity)
	//		transfer.TransferEntity = new(models.TransferData)
	//		transferKey = NewTransferKey(c, transferID)
	//		keys := []*datastore.Key{receiptKey, transferKey}
	//		if err = gaedb.GetMulti(c, keys, []interface{}{receipt.ReceiptEntity, transfer.TransferEntity}); err != nil {
	//			return err
	//		}
	//	}
	//
	//	if receipt.DtSent.IsZero() {
	//		receipt.DtSent = sentTime
	//		isReceiptIdIsInTransfer := false
	//		for _, rId := range transfer.ReceiptIDs {
	//			if rId == receiptID {
	//				isReceiptIdIsInTransfer = true
	//				break
	//			}
	//		}
	//		if isReceiptIdIsInTransfer {
	//			_, err = gaedb.Put(c, receiptKey, receipt)
	//		} else {
	//			transfer.ReceiptIDs = append(transfer.ReceiptIDs, receiptID)
	//			transfer.ReceiptsSentCount += 1
	//			_, err = gaedb.PutMulti(c, []*datastore.Key{receiptKey, transferKey}, []interface{}{receipt.ReceiptEntity, transfer.TransferEntity})
	//		}
	//	}
	//	return err
	//}, dtdal.CrossGroupTransaction)
}

func (receiptDalGae ReceiptDalGae) DelayedMarkReceiptAsSent(c context.Context, receiptID, transferID int, sentTime time.Time) error {
	return gae.CallDelayFunc(c, common.QUEUE_TRANSFERS, "set-receipt-as-sent", delayedMarkReceiptAsSent, receiptID, transferID, sentTime)
}

var delayedMarkReceiptAsSent = delay.MustRegister("delayedMarkReceiptAsSent", func(c context.Context, receiptID, transferID int, sentTime time.Time) (err error) {
	log.Debugf(c, "delayedMarkReceiptAsSent(receiptID=%v, transferID=%v, sentTime=%v)", receiptID, transferID, sentTime)
	if receiptID == 0 {
		log.Errorf(c, "receiptID == 0")
		return nil
	}
	if receiptID == 0 {
		log.Errorf(c, "transferID == 0")
		return nil
	}

	if err = dtdal.Receipt.MarkReceiptAsSent(c, receiptID, transferID, sentTime); dal.IsNotFound(err) {
		log.Errorf(c, err.Error())
		return nil
	}
	return
})
