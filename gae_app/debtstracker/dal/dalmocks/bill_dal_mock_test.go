package dalmocks

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"golang.org/x/net/context"
)

var _ dal.BillDal = (*BillDalMock)(nil)

type BillDalMock struct {
	Bills map[string]*models.BillEntity
}

func NewBillDalMock() *BillDalMock {
	return &BillDalMock{Bills: make(map[string]*models.BillEntity)}
}

func (billDalMock *BillDalMock) InsertBillEntity(c context.Context, billEntity *models.BillEntity) (bill models.Bill, err error) {
	bill.ID = db.RandomStringID(8)
	billDalMock.Bills[bill.ID] = billEntity
	bill.BillEntity = billEntity
	return
}

func (billDalMock *BillDalMock) UpdateBill(c context.Context, bill models.Bill) (err error) {
	billDalMock.Bills[bill.ID] = bill.BillEntity
	return
}

func (billDalMock *BillDalMock) GetBillByID(c context.Context, billID string) (bill models.Bill, err error) {
	if billEntity, ok := billDalMock.Bills[billID]; !ok {
		err = db.NewErrNotFoundByStrID(models.BillKind, billID, errors.New("Not found"))
	} else {
		bill = models.Bill{StringID: db.StringID{ID: billID}, BillEntity: billEntity}
	}
	return
}

func (billDalMock *BillDalMock) UpdateBillsHolder(c context.Context, billID string, getBillsHolder dal.BillsHolderGetter) (err error) {
	panic(NOT_IMPLEMENTED_YET)
}
