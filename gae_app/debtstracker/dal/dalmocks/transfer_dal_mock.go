package dalmocks

import (
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/mockdb"
)

const NOT_IMPLEMENTED_YET = "NOT_IMPLEMENTED_YET"

type TransferDalMock struct {
	mockDB *mockdb.MockDB
}

func NewTransferDalMock(mockDB *mockdb.MockDB) *TransferDalMock {
	return &TransferDalMock{
		mockDB: mockDB,
	}
}

func (mock *TransferDalMock) DelayUpdateTransfersOnReturn(c context.Context, returnTransferID int64, transferReturnUpdates []dal.TransferReturnUpdate) (err error) {
	panic("not implemented yet")
}

func (mock *TransferDalMock) GetTransfersByID(c context.Context, transferIDs []int64) ([]models.Transfer, error) {
	panic("not implemented yet")
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
	for _, entity := range mock.mockDB.EntitiesByKind[models.TransferKind] {
		t := entity.(*models.Transfer)
		if t.GetOutstandingValue(periodEnds) != 0 {
			transfers = append(transfers, *t)
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
