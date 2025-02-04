package debtusdal

import (
	"context"
	"errors"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal/delayer4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/const4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/dal4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/strongo/delaying"
	"github.com/strongo/logus"
	"time"
)

type ReceiptDalGae struct {
}

func NewReceiptDalGae() ReceiptDalGae {
	return ReceiptDalGae{}
}

var _ dtdal.ReceiptDal = (*ReceiptDalGae)(nil)

func (ReceiptDalGae) DelayCreateAndSendReceiptToCounterpartyByTelegram(ctx context.Context, env string, transferID string, userID string) error {
	logus.Debugf(ctx, "delayerSendReceiptToCounterpartyByTelegram(env=%v, transferID=%v, userID=%v)", env, transferID, userID)
	return delayer4debtus.CreateAndSendReceiptToCounterpartyByTelegram.EnqueueWork(ctx, delaying.With(const4debtus.QueueReceipts, "create-and-send-receipt-for-counterparty-by-telegram", 0), env, transferID, userID)
}

func (ReceiptDalGae) UpdateReceipt(ctx context.Context, tx dal.ReadwriteTransaction, receipt models4debtus.ReceiptEntry) error {
	return tx.Set(ctx, receipt.Record)
}

func (receiptDalGae ReceiptDalGae) GetReceiptByID(ctx context.Context, tx dal.ReadSession, id string) (receipt models4debtus.ReceiptEntry, err error) {
	receipt = models4debtus.NewReceipt(id, nil)
	return receipt, tx.Get(ctx, receipt.Record)
}

func (receiptDalGae ReceiptDalGae) CreateReceipt(ctx context.Context, data *models4debtus.ReceiptDbo) (receipt models4debtus.ReceiptEntry, err error) { // TODO: Move to facade4debtus
	err = facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		receipt = models4debtus.NewReceiptWithoutID(data)
		debtusUser := models4debtus.NewDebtusUserEntry(data.CreatorUserID)
		if err = dal4debtus.GetDebtusUser(ctx, tx, debtusUser); err != nil {
			return err
		}
		debtusUser.Data.CountOfReceiptsCreated += 1
		if err = tx.Set(ctx, debtusUser.Record); err != nil {
			return err
		}
		if err = tx.Insert(ctx, receipt.Record); err != nil {
			return err
		}
		receipt.ID = receipt.Record.Key().ID.(string)
		return
	})
	return
}

func (receiptDalGae ReceiptDalGae) MarkReceiptAsSent(ctx context.Context, receiptID, transferID string, sentTime time.Time) error {
	return errors.New("TODO: Implement MarkReceiptAsSent")
	//return dtdal.DB.RunInTransaction(ctx, func(ctx context.Context) (err error) {
	//	var (
	//		receipt     models.ReceiptEntry
	//		transfer    models.TransferEntry
	//		transferKey *datastore.Key
	//	)
	//	receiptKey := NewReceiptKey(ctx, receiptID)
	//	if transferID == 0 {
	//		if receipt, err = receiptDalGae.GetReceiptByID(ctx, receiptID); err != nil {
	//			return err
	//		}
	//		if transfer, err = facade4debtus.QueueTransfers.GetTransferByID(ctx, transferID); err != nil {
	//			return err
	//		}
	//		transferKey = NewTransferKey(ctx, transferID)
	//	} else {
	//		receipt.ReceiptDbo = new(models.ReceiptDbo)
	//		transfer.TransferEntity = new(models.TransferData)
	//		transferKey = NewTransferKey(ctx, transferID)
	//		keys := []*datastore.Key{receiptKey, transferKey}
	//		if err = gaedb.GetMulti(ctx, keys, []interface{}{receipt.ReceiptDbo, transfer.TransferEntity}); err != nil {
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
	//			_, err = gaedb.Put(ctx, receiptKey, receipt)
	//		} else {
	//			transfer.ReceiptIDs = append(transfer.ReceiptIDs, receiptID)
	//			transfer.ReceiptsSentCount += 1
	//			_, err = gaedb.PutMulti(ctx, []*datastore.Key{receiptKey, transferKey}, []interface{}{receipt.ReceiptDbo, transfer.TransferEntity})
	//		}
	//	}
	//	return err
	//}, dtdal.CrossGroupTransaction)
}

func (receiptDalGae ReceiptDalGae) DelayedMarkReceiptAsSent(ctx context.Context, receiptID, transferID string, sentTime time.Time) error {
	return delayer4debtus.MarkReceiptAsSent.EnqueueWork(ctx, delaying.With(const4debtus.QueueTransfers, "set-receipt-as-sent", 0), receiptID, transferID, sentTime)
}

func delayedMarkReceiptAsSent(ctx context.Context, receiptID, transferID string, sentTime time.Time) (err error) {
	logus.Debugf(ctx, "MarkReceiptAsSent(receiptID=%v, transferID=%v, sentTime=%v)", receiptID, transferID, sentTime)
	if receiptID == "" {
		logus.Errorf(ctx, "receiptID == 0")
		return nil
	}
	if receiptID == "" {
		logus.Errorf(ctx, "transferID == 0")
		return nil
	}

	if err = dtdal.Receipt.MarkReceiptAsSent(ctx, receiptID, transferID, sentTime); dal.IsNotFound(err) {
		logus.Errorf(ctx, err.Error())
		return nil
	}
	return
}
