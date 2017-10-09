package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/app/db"
)

func NewTransferKey(c context.Context, transferID int64) *datastore.Key {
	if transferID == 0 {
		panic("transferID == 0")
	}
	return gaedb.NewKey(c, models.TransferKind, "", transferID, nil)
}

func NewTransferIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.TransferKind, nil)
}

type TransferDalGae struct {
}

func NewTransferDalGae() TransferDalGae {
	return TransferDalGae{}
}

var _ dal.TransferDal = (*TransferDalGae)(nil)

func _loadDueOnTransfers(c context.Context, userID int64, limit int, filter func(q *datastore.Query) *datastore.Query) (transfers []models.Transfer, err error) {
	q := datastore.NewQuery(models.TransferKind)
	q = filter(q.Filter("BothUserIDs =", userID).Filter("IsOutstanding =", true))
	q = q.Order("DtDueOn")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var (
		transferKeys     []*datastore.Key
		transferEntities []*models.TransferEntity
	)

	if transferKeys, err = q.GetAll(c, &transferEntities); err != nil {
		return
	}
	transfers = make([]models.Transfer, len(transferKeys))
	for i, transferKey := range transferKeys {
		transfer := models.NewTransfer(transferKey.IntID(), transferEntities[i])
		transfers[i] = transfer
	}
	return
}

func (transferDalGae TransferDalGae) LoadOverdueTransfers(c context.Context, userID int64, limit int) ([]models.Transfer, error) {
	return _loadDueOnTransfers(c, userID, limit, func(q *datastore.Query) *datastore.Query {
		return q.Filter("DtDueOn >", time.Time{}).Filter("DtDueOn <", time.Now())
	})
}

func (transferDalGae TransferDalGae) LoadDueTransfers(c context.Context, userID int64, limit int) ([]models.Transfer, error) {
	return _loadDueOnTransfers(c, userID, limit, func(q *datastore.Query) *datastore.Query {
		return q.Filter("DtDueOn >", time.Now())
	})
}

func (transferDalGae TransferDalGae) GetTransferByID(c context.Context, id int64) (models.Transfer, error) {
	var transferEntity models.TransferEntity
	key := NewTransferKey(c, id)
	if err := gaedb.Get(c, key, &transferEntity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.NewErrNotFoundByIntID(models.TransferKind, id, err)
		} else {
			err = errors.Wrapf(err, "Failed to get transfer by id=%v", id)
		}
		return models.Transfer{ID: id}, err
	}
	return models.NewTransfer(id, &transferEntity), nil
}

func (transferDalGae TransferDalGae) InsertTransfer(c context.Context, transferEntity *models.TransferEntity) (transfer models.Transfer, err error) {
	log.Debugf(c, "TransferDalGae.InsertTransfer(%v)", *transferEntity)
	key := NewTransferIncompleteKey(c)
	if key, err = gaedb.Put(c, key, transferEntity); err != nil {
		err = errors.Wrap(err, "Failed to insert transfer")
		return
	}
	transfer = models.NewTransfer(key.IntID(), transferEntity)
	return
}

func (transferDalGae TransferDalGae) SaveTransfer(c context.Context, transfer models.Transfer) error {
	if transfer.ID == 0 {
		panic("transfer.ID == 0")
	}
	if _, err := gaedb.Put(c, NewTransferKey(c, transfer.ID), transfer.TransferEntity); err != nil {
		return errors.Wrap(err, "Failed to save transfer")
	} else {
		return nil
	}
}

func (transferDalGae TransferDalGae) LoadOutstandingTransfers(c context.Context, userID int64, currency models.Currency) (transfers []models.Transfer, err error) {
	const limit = 100
	q := datastore.NewQuery(models.TransferKind)
	q = q.Filter("BothUserIDs =", userID)
	q = q.Filter("Currency =", string(currency))
	q = q.Filter("IsOutstanding =", true)
	q = q.Order("DtCreated")
	q = q.Limit(limit)
	transferEntities := make([]*models.TransferEntity, 0, limit)
	var keys []*datastore.Key
	if keys, err = q.GetAll(c, &transferEntities); err != nil {
		return
	}
	transfers = make([]models.Transfer, len(keys))
	for i, key := range keys {
		transfers[i] = models.Transfer{ID: key.IntID(), TransferEntity: transferEntities[i]}
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
	q := datastore.NewQuery(models.TransferKind)
	q = q.Filter("BothUserIDs =", userID)
	q = q.Offset(offset)
	q = q.Order("-DtCreated")
	q = q.Limit(limit + 1)

	if transfers, err = transferDalGae.loadTransfers(c, q); err != nil {
		return
	}
	hasMore = len(transfers) > limit
	return
}

func (transferDalGae TransferDalGae) LoadTransferIDsByContactID(c context.Context, contactID int64, limit int, startCursor string) (transferIDs []int64, endCursor string, err error) {
	if limit == 0 {
		err = errors.New("limit == 0")
		return
	} else if limit > 1000 {
		err = errors.New("limit > 1000")
		return
	}
	if contactID == 0 {
		err = errors.New("contactID == 0")
		return
	}
	q := datastore.NewQuery(models.TransferKind)
	q = q.Filter("BothCounterpartyIDs =", contactID)
	q = q.Limit(limit + 1)
	q = q.KeysOnly()
	if startCursor != "" {
		var decodedCursor datastore.Cursor
		if decodedCursor, err = datastore.DecodeCursor(startCursor); err != nil {
			return
		} else {
			q = q.Start(decodedCursor)
		}
	}

	var key *datastore.Key
	transferIDs = make([]int64, 0, limit)
	for t := q.Run(c); ; {
		key, err = t.Next(nil)
		if err == datastore.Done {
			err = nil
			var cursor datastore.Cursor
			if cursor, err = t.Cursor(); err != nil {
				return
			}
			endCursor = cursor.String()
			return
		} else if err != nil {
			return
		}
		transferIDs = append(transferIDs, key.IntID())
	}
}

func (transferDalGae TransferDalGae) LoadTransfersByContactID(c context.Context, contactID int64, offset, limit int) (transfers []models.Transfer, hasMore bool, err error) {
	if limit == 0 {
		err = errors.New("limit == 0")
		return
	}
	if contactID == 0 {
		err = errors.New("contactID == 0")
		return
	}
	q := datastore.NewQuery(models.TransferKind)
	q = q.Filter("BothCounterpartyIDs =", contactID)
	q = q.Offset(offset)
	q = q.Order("-DtCreated")
	q = q.Limit(limit + 1)

	if transfers, err = transferDalGae.loadTransfers(c, q); err != nil {
		return
	}
	hasMore = len(transfers) > limit
	return
}

func (transferDalGae TransferDalGae) LoadLatestTransfers(c context.Context, offset, limit int) ([]models.Transfer, error) {
	q := datastore.NewQuery(models.TransferKind)
	q = q.Offset(offset)
	q = q.Order("-DtCreated")
	q = q.Limit(limit)

	return transferDalGae.loadTransfers(c, q)
}

func (transferDalGae TransferDalGae) loadTransfers(c context.Context, q *datastore.Query) (transfers []models.Transfer, err error) {
	var (
		transferKeys     []*datastore.Key
		transferEntities []*models.TransferEntity
	)
	if transferKeys, err = q.GetAll(c, &transferEntities); err != nil {
		err = errors.Wrap(err, "Failed to loadTransfers()")
		return
	}
	log.Debugf(c, "loadTransfers(): %v", transferKeys)
	transfers = make([]models.Transfer, len(transferKeys))
	for i, transferKey := range transferKeys {
		transferEntity := transferEntities[i]
		transfers[i] = models.Transfer{
			ID:             transferKey.IntID(),
			TransferEntity: transferEntity,
		}
	}
	return
}
