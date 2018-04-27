package dalmocks

import (
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
)

type TransferDalMock struct {
	Transfers map[int64]*models.TransferEntity
}

func NewTransferDalMock() *TransferDalMock {
	return &TransferDalMock{
		Transfers: make(map[int64]*models.TransferEntity),
	}
}

func (mock *TransferDalMock) DelayUpdateTransfersOnReturn(c context.Context, returnTransferID int64, transferReturnUpdates []dal.TransferReturnUpdate) (err error) {
	panic("not implemented yet")
}

func (mock *TransferDalMock) GetTransferByID(c context.Context, transferID int64) (models.Transfer, error) {
	if transferEntity, ok := mock.Transfers[transferID]; ok {
		return models.Transfer{IntegerID: db.NewIntID(transferID), TransferEntity: transferEntity}, nil
	} else {
		return models.Transfer{}, db.NewErrNotFoundByIntID(models.TransferKind, transferID, nil)
	}
}

func (mock *TransferDalMock) GetTransfersByID(c context.Context, transferIDs []int64) ([]models.Transfer, error) {
	panic("not implemented yet")
}

func (mock *TransferDalMock) UpdateTransferOnReturn(c context.Context, returnTransfer, transfer models.Transfer, returnedAmount decimal.Decimal64p2) (err error) {
	panic("not implemented yet")
}

func (mock *TransferDalMock) SaveTransfer(c context.Context, transfer models.Transfer) error {
	if _, err := transfer.TransferEntity.Save(); err != nil {
		return err
	}
	mock.Transfers[transfer.ID] = transfer.TransferEntity
	return nil
}

func (mock *TransferDalMock) InsertTransfer(c context.Context, transferEntity *models.TransferEntity) (transfer models.Transfer, err error) {
	if transferEntity == nil {
		panic("transferEntity == nil")
	}
	var maxTransferID int64
	for transferID := range mock.Transfers {
		if transferID > maxTransferID {
			maxTransferID = transferID
		}
	}
	maxTransferID += 1
	if _, err = transferEntity.Save(); err != nil {
		return
	}
	mock.Transfers[maxTransferID] = transferEntity
	return models.NewTransfer(maxTransferID, transferEntity), nil
}

func (mock *TransferDalMock) LoadTransfersByUserID(c context.Context, userID int64, offset, limit int) (transfers []models.Transfer, hasMore bool, err error) {
	panic(NOT_IMPLEMENTED_YET)
}

func (mock *TransferDalMock) LoadTransferIDsByContactID(c context.Context, contactID int64, limit int, startCursor string) (transferIDs []int64, endCursor string, err error) {
	panic(NOT_IMPLEMENTED_YET)
}

func (mock *TransferDalMock) LoadTransfersByContactID(c context.Context, contactID int64, offset, limit int) (transfers []models.Transfer, hasMore bool, err error) {
	panic(NOT_IMPLEMENTED_YET)
}

func (mock *TransferDalMock) LoadOverdueTransfers(c context.Context, userID int64, limit int) (transfers []models.Transfer, err error) {
	panic(NOT_IMPLEMENTED_YET)
}

func (mock *TransferDalMock) LoadOutstandingTransfers(c context.Context, periodEnds time.Time, userID, contactID int64, currency models.Currency, direction models.TransferDirection) (transfers []models.Transfer, err error) {
	for id, t := range mock.Transfers {
		if t.GetOutstandingValue(periodEnds) != 0 {
			transfers = append(transfers, models.Transfer{IntegerID: db.NewIntID(id), TransferEntity: t})
		}
	}
	return
}

func (mock *TransferDalMock) LoadDueTransfers(c context.Context, userID int64, limit int) (transfers []models.Transfer, err error) {
	panic(NOT_IMPLEMENTED_YET)
}

func (mock *TransferDalMock) LoadLatestTransfers(c context.Context, offset, limit int) ([]models.Transfer, error) {
	panic(NOT_IMPLEMENTED_YET)
}

func (mock *TransferDalMock) DelayUpdateTransferWithCreatorReceiptTgMessageID(c context.Context, botCode string, transferID, creatorTgChatID, creatorTgReceiptMessageID int64) error {
	panic(NOT_IMPLEMENTED_YET)
}

func (mock *TransferDalMock) DelayUpdateTransfersWithCounterparty(c context.Context, creatorCounterpartyID, counterpartyCounterpartyID int64) error {
	panic(NOT_IMPLEMENTED_YET)
}
