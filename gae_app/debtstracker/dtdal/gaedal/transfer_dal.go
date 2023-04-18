package gaedal

import (
	"bytes"
	"fmt"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
)

type TransferDalGae struct {
}

func NewTransferDalGae() TransferDalGae {
	return TransferDalGae{}
}

var _ dtdal.TransferDal = (*TransferDalGae)(nil)

func _loadDueOnTransfers(c context.Context, tx dal.ReadSession, userID int64, limit int, filter func(q dal.Selector) dal.Selector) (transfers []models.Transfer, err error) {
	q := dal.From(models.TransferKind).
		WhereField("BothUserIDs", "=", userID).
		WhereField("IsOutstanding", "=", true).OrderBy(dal.AscendingField("DtDueOn"))
	q = filter(q)
	query := q.SelectInto(models.NewTransferRecord)
	if limit > 0 {
		query.Limit = limit
	}
	var (
		transferRecords []dal.Record
	)

	transferRecords, err = tx.SelectAll(c, query)

	transfers = make([]models.Transfer, len(transferRecords))
	for i, transferRecord := range transferRecords {
		transfer := models.NewTransfer(transferRecord.Key().ID.(int), transferRecord.Data().(*models.TransferData))
		transfers[i] = transfer
	}
	return
}

func (transferDalGae TransferDalGae) LoadOverdueTransfers(c context.Context, tx dal.ReadSession, userID int64, limit int) ([]models.Transfer, error) {
	return _loadDueOnTransfers(c, tx, userID, limit, func(q dal.Selector) dal.Selector {
		return q.WhereField("DtDueOn", dal.GreaterThen, time.Time{}).WhereField("DtDueOn", dal.LessThen, time.Now())
	})
}

func (transferDalGae TransferDalGae) LoadDueTransfers(c context.Context, tx dal.ReadSession, userID int64, limit int) ([]models.Transfer, error) {
	return _loadDueOnTransfers(c, tx, userID, limit, func(q dal.Selector) dal.Selector {
		return q.WhereField("DtDueOn", dal.GreaterThen, time.Now())
	})
}

func (transferDalGae TransferDalGae) GetTransfersByID(c context.Context, tx dal.ReadSession, transferIDs []int) (transfers []models.Transfer, err error) {
	transfers = make([]models.Transfer, len(transferIDs))
	records := make([]dal.Record, len(transferIDs))
	for i, transferID := range transferIDs {
		transfers[i] = models.NewTransfer(transferID, nil)
		records[i] = transfers[i].Record
	}
	if err = tx.GetMulti(c, records); err != nil {
		return
	}
	return
}

func (transferDalGae TransferDalGae) LoadOutstandingTransfers(c context.Context, tx dal.ReadSession, periodEnds time.Time, userID, contactID int64, currency money.Currency, direction models.TransferDirection) (transfers []models.Transfer, err error) {
	log.Debugf(c, "TransferDalGae.LoadOutstandingTransfers(periodEnds=%v, userID=%v, contactID=%v currency=%v, direction=%v)", periodEnds, userID, contactID, currency, direction)
	const limit = 100

	// TODO: Load outstanding transfer just for the specific contact & specific direction
	q := dal.From(models.TransferKind).
		Where(
			dal.WhereField("BothUserIDs", dal.Equal, userID),
			dal.WhereField("Currency", dal.Equal, string(currency)),
			dal.WhereField("IsOutstanding", dal.Equal, true),
		).OrderBy(dal.AscendingField("DtCreated")).
		SelectInto(models.NewTransferRecord)
	q.Limit = limit
	var transferRecords []dal.Record
	transferRecords, err = tx.SelectAll(c, q)
	transfers = models.TransfersFromRecords(transferRecords)
	var errorMessages, warnings, debugs bytes.Buffer
	var transfersIDsToFixIsOutstanding []int
	for _, transfer := range transfers {
		if contactID != 0 {
			if cpContactID := transfer.Data.CounterpartyInfoByUserID(userID).ContactID; cpContactID != contactID {
				debugs.WriteString(fmt.Sprintf("Skipped outstanding Transfer(id=%v) as counterpartyContactID != contactID: %v != %v\n", transfer.ID, cpContactID, contactID))
				continue
			}
		}
		if direction != "" {
			if d := transfer.Data.DirectionForUser(userID); d != direction {
				debugs.WriteString(fmt.Sprintf("Skipped outstanding Transfer(id=%v) as DirectionForUser(): %v\n", transfer.ID, d))
				continue
			}
		}

		if outstandingValue := transfer.Data.GetOutstandingValue(periodEnds); outstandingValue > 0 {
			transfers = append(transfers, transfer)
		} else if outstandingValue == 0 {
			_, _ = fmt.Fprintf(&warnings, "Transfer(id=%v) => GetOutstandingValue() == 0 && IsOutstanding==true\n", transfer.ID)
			transfersIDsToFixIsOutstanding = append(transfersIDsToFixIsOutstanding, transfer.ID)
		} else { // outstandingValue < 0
			_, _ = fmt.Fprintf(&errorMessages, "Transfer(id=%v) => IsOutstanding==true && GetOutstandingValue() < 0: %v\n", transfer.ID, outstandingValue)
		}
	}
	if len(transfersIDsToFixIsOutstanding) > 0 {
		if err = gae.CallDelayFunc(c, common.QUEUE_TRANSFERS, "fix-transfers-is-outstanding", delayFixTransfersIsOutstanding, transfersIDsToFixIsOutstanding); err != nil {
			log.Errorf(c, "failed to delay task to fix transfers IsOutstanding")
			err = nil
		}
	}
	if errorMessages.Len() > 0 {
		log.Errorf(c, errorMessages.String())
	}
	if warnings.Len() > 0 {
		log.Warningf(c, warnings.String())
	}
	if debugs.Len() > 0 {
		log.Debugf(c, debugs.String())
	}
	return
}

var delayFixTransfersIsOutstanding = delay.Func("fix-transfers-is-outstanding", fixTransfersIsOutstanding)

func fixTransfersIsOutstanding(c context.Context, transferIDs []int) (err error) {
	log.Debugf(c, "fixTransfersIsOutstanding(%v)", transferIDs)
	for _, transferID := range transferIDs {
		if _, transferErr := fixTransferIsOutstanding(c, transferID); transferErr != nil {
			log.Errorf(c, "Failed to fix transfer %v: %v", transferID, err)
			err = transferErr
		}
	}
	return
}

func fixTransferIsOutstanding(c context.Context, transferID int) (transfer models.Transfer, err error) {
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		if transfer, err = facade.Transfers.GetTransferByID(c, tx, transferID); err != nil {
			return err
		}
		if transfer.Data.GetOutstandingValue(time.Now()) == 0 {
			transfer.Data.IsOutstanding = false
			return facade.Transfers.SaveTransfer(c, tx, transfer)
		}
		return nil
	})
	if err == nil {
		log.Warningf(c, "Fixed IsOutstanding (set to false) for transfer %v", transferID)
	} else {
		log.Errorf(c, "Failed to fix IsOutstanding for transfer %v", transferID)
	}
	return
}

func (transferDalGae TransferDalGae) LoadTransfersByUserID(c context.Context, userID int64, offset, limit int) (transfers []models.Transfer, hasMore bool, err error) {
	if limit == 0 {
		err = errors.New("limit == 0")
		return
	}
	if userID == 0 {
		err = errors.New("userID == 0")
		return
	}
	q := dal.From(models.TransferKind).
		WhereField("BothUserIDs", dal.Equal, userID).
		OrderBy(dal.DescendingField("DtCreated")).
		SelectInto(models.NewTransferRecord)

	if transfers, err = transferDalGae.loadTransfers(c, q); err != nil {
		return
	}
	hasMore = len(transfers) > limit
	return
}

func (transferDalGae TransferDalGae) LoadTransferIDsByContactID(c context.Context, contactID int64, limit int, startCursor string) (transferIDs []int, endCursor string, err error) {
	if limit == 0 {
		err = errors.New("LoadTransferIDsByContactID(): limit == 0")
		return
	} else if limit > 1000 {
		err = errors.New("LoadTransferIDsByContactID(): limit > 1000")
		return
	}
	if contactID == 0 {
		err = errors.New("LoadTransferIDsByContactID(): contactID == 0")
		return
	}
	q := dal.From(models.TransferKind).
		WhereField("BothCounterpartyIDs", dal.Equal, contactID).
		SelectInto(models.NewTransferRecord)
	q.Limit = limit
	q.StartCursor = startCursor

	//if startCursor != "" {
	//	var decodedCursor datastore.Cursor
	//	if decodedCursor, err = datastore.DecodeCursor(startCursor); err != nil {
	//		return
	//	} else {
	//		q = q.Start(decodedCursor)
	//	}
	//}

	transferIDs = make([]int, 0, limit)
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	var reader dal.Reader
	if reader, err = db.Select(c, q); err != nil {
		return
	}
	var record dal.Record
	for record, err = reader.Next(); err != nil; {
		if dal.ErrNoMoreRecords == err {
			if endCursor, err = reader.Cursor(); err != nil {
				return
			}
			return
		} else if err != nil {
			return
		}
		transferIDs = append(transferIDs, record.Key().ID.(int))
	}
	return
}

func (transferDalGae TransferDalGae) LoadTransfersByContactID(c context.Context, contactID int64, offset, limit int) (transfers []models.Transfer, hasMore bool, err error) {
	if limit == 0 {
		err = errors.New("LoadTransfersByContactID(): limit == 0")
		return
	}
	if contactID == 0 {
		err = errors.New("LoadTransfersByContactID(): contactID == 0")
		return
	}
	q := dal.From(models.TransferKind).
		WhereField("BothCounterpartyIDs", dal.Equal, contactID).
		OrderBy(dal.DescendingField("DtCreated")).
		SelectInto(models.NewTransferRecord)
	q.Limit = limit
	q.Offset = offset

	if transfers, err = transferDalGae.loadTransfers(c, q); err != nil {
		return
	}
	hasMore = len(transfers) > limit
	return
}

func (transferDalGae TransferDalGae) LoadLatestTransfers(c context.Context, offset, limit int) ([]models.Transfer, error) {
	q := dal.From(models.TransferKind).
		OrderBy(dal.DescendingField("DtCreated")).
		SelectInto(models.NewTransferRecord)
	q.Limit = limit
	q.Offset = offset
	return transferDalGae.loadTransfers(c, q)
}

func (transferDalGae TransferDalGae) loadTransfers(c context.Context, q dal.Query) (transfers []models.Transfer, err error) {
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	var records []dal.Record
	if records, err = db.SelectAll(c, q); err != nil {
		return
	}
	transfers = make([]models.Transfer, len(records))
	for i, record := range records {
		transfers[i] = models.NewTransfer(record.Key().ID.(int), record.Data().(*models.TransferData))
	}
	return transfers, nil
}
