package gaedal

import (
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
)

func NewReceiptKey(c context.Context, id int64) *datastore.Key {
	return gaedb.NewKey(c, models.ReceiptKind, "", id, nil)
}

func NewReceiptIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.ReceiptKind, nil)
}

type ReceiptDalGae struct {
}

func NewReceiptDalGae() ReceiptDalGae {
	return ReceiptDalGae{}
}

var _ dal.ReceiptDal = (*ReceiptDalGae)(nil)

func (ReceiptDalGae) UpdateReceipt(c context.Context, receipt models.Receipt) error {
	_, err := gaedb.Put(c, NewReceiptKey(c, receipt.ID), receipt.ReceiptEntity)
	return err
}

func (receiptDalGae ReceiptDalGae) GetReceiptByID(c context.Context, id int64) (models.Receipt, error) {
	receiptEntity := new(models.ReceiptEntity)
	err := gaedb.Get(c, gaedb.NewKey(c, models.ReceiptKind, "", id, nil), receiptEntity)
	if err == datastore.ErrNoSuchEntity {
		err = db.NewErrNotFoundByIntID(models.ReceiptKind, id, err)
	} else if err != nil {
		err = errors.Wrapf(err, "Failed to get receipt by id=%v", id)
	}
	return models.Receipt{IntegerID: db.NewIntID(id), ReceiptEntity: receiptEntity}, err
}

func (receiptDalGae ReceiptDalGae) CreateReceipt(c context.Context, receipt *models.ReceiptEntity) (id int64, err error) { // TODO: Move to facade
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		receiptKey := NewReceiptIncompleteKey(c)
		user, err := dal.User.GetUserByID(c, receipt.CreatorUserID)
		if err != nil {
			return err
		}
		user.CountOfReceiptsCreated += 1
		if keys, err := gaedb.PutMulti(c, []*datastore.Key{receiptKey, NewAppUserKey(c, receipt.CreatorUserID)}, []interface{}{receipt, user}); err != nil {
			return err
		} else {
			receiptKey = keys[0]
		}
		id = receiptKey.IntID()
		return nil
	}, dal.CrossGroupTransaction)
	return
}

func (receiptDalGae ReceiptDalGae) MarkReceiptAsSent(c context.Context, receiptID, transferID int64, sentTime time.Time) error {
	return dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var (
			receipt     models.Receipt
			transfer    models.Transfer
			transferKey *datastore.Key
		)
		receiptKey := NewReceiptKey(c, receiptID)
		if transferID == 0 {
			if receipt, err = receiptDalGae.GetReceiptByID(c, receiptID); err != nil {
				return err
			}
			if transfer, err = dal.Transfer.GetTransferByID(c, transferID); err != nil {
				return err
			}
			transferKey = NewTransferKey(c, transferID)
		} else {
			receipt.ReceiptEntity = new(models.ReceiptEntity)
			transfer.TransferEntity = new(models.TransferEntity)
			transferKey = NewTransferKey(c, transferID)
			keys := []*datastore.Key{receiptKey, transferKey}
			if err = gaedb.GetMulti(c, keys, []interface{}{receipt.ReceiptEntity, transfer.TransferEntity}); err != nil {
				return err
			}
		}

		if receipt.DtSent.IsZero() {
			receipt.DtSent = sentTime
			isReceiptIdIsInTransfer := false
			for _, rId := range transfer.ReceiptIDs {
				if rId == receiptID {
					isReceiptIdIsInTransfer = true
					break
				}
			}
			if isReceiptIdIsInTransfer {
				_, err = gaedb.Put(c, receiptKey, receipt)
			} else {
				transfer.ReceiptIDs = append(transfer.ReceiptIDs, receiptID)
				transfer.ReceiptsSentCount += 1
				_, err = gaedb.PutMulti(c, []*datastore.Key{receiptKey, transferKey}, []interface{}{receipt.ReceiptEntity, transfer.TransferEntity})
			}
		}
		return err
	}, dal.CrossGroupTransaction)
}

func (receiptDalGae ReceiptDalGae) DelayedMarkReceiptAsSent(c context.Context, receiptID, transferID int64, sentTime time.Time) error {
	return gae.CallDelayFunc(c, common.QUEUE_TRANSFERS, "set-receipt-as-sent", delayedMarkReceiptAsSent, receiptID, transferID, sentTime)
}

var delayedMarkReceiptAsSent = delay.Func("delayedMarkReceiptAsSent", func(c context.Context, receiptID, transferID int64, sentTime time.Time) (err error) {
	log.Debugf(c, "delayedMarkReceiptAsSent(receiptID=%v, transferID=%v, sentTime=%v)", receiptID, transferID, sentTime)
	if receiptID == 0 {
		log.Errorf(c, "receiptID == 0")
		return nil
	}
	if receiptID == 0 {
		log.Errorf(c, "transferID == 0")
		return nil
	}

	if err = dal.Receipt.MarkReceiptAsSent(c, receiptID, transferID, sentTime); db.IsNotFound(err) {
		log.Errorf(c, err.Error())
		return nil
	}
	return
})
