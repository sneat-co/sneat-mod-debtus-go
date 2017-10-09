package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/delay"
	"time"
)

type billDalGae struct {
}

var _ dal.BillDal = (*billDalGae)(nil) // Make sure we implement interface

func newBillDalGae() billDalGae {
	return billDalGae{}
}

func (_ billDalGae) GetBillByID(c context.Context, billID string) (bill models.Bill, err error) {
	bill.ID = billID
	return bill, dal.DB.Get(c, &bill)
}

func (_ billDalGae) GetBillsByIDs(c context.Context, billIDs []string) (bills []models.Bill, err error) {
	entityHolders := make([]db.EntityHolder, len(billIDs))
	for i, id := range billIDs {
		entityHolders[i] = &models.Bill{StringID: db.StringID{ID: id}}
	}
	bills = make([]models.Bill, len(entityHolders))
	if err = dal.DB.GetMulti(c, entityHolders); err != nil {
		return
	}
	for i, bill := range entityHolders {
		bills[i] = *bill.(*models.Bill)
	}
	return
}

func (_ billDalGae) InsertBillEntity(c context.Context, billEntity *models.BillEntity) (bill models.Bill, err error) {
	if billEntity.CreatorUserID == 0 {
		panic("CreatorUserID == 0")
	}
	if billEntity.AmountTotal == 0 {
		panic("AmountTotal == 0")
	}

	billEntity.DtCreated = time.Now()

	err = dal.InsertWithRandomStringID(c, &bill, 8)
	return
}

func (billDalGae) UpdateBill(c context.Context, bill models.Bill) (err error) {
	return dal.DB.Update(c, &bill)
}

func (billDalGae) DelayUpdateBillDependencies(c context.Context, billID string) (err error) {
	if err = gae.CallDelayFunc(c, common.QUEUE_BILLS, "UpdateBillDependencies", delayedUpdateBillDependencies, billID); err != nil {
		return
	}
	return
}

var delayedUpdateBillDependencies = delay.Func("delayedUpdateBillDependencies", func(c context.Context, billID string) (err error) {
	log.Debugf(c, "delayedUpdateBillDependencies(billID=%d)", billID)
	var bill models.Bill
	if bill, err = dal.Bill.GetBillByID(c, billID); err != nil {
		if db.IsNotFound(err) {
			log.Warningf(c, err.Error())
			err = nil
		}
		return
	}
	for _, groupID := range bill.UserGroupIDs {
		if err = dal.Group.DelayUpdateGroupWithBill(c, groupID, bill.ID); err != nil {
			return
		}
	}
	for _, member := range bill.GetBillMembers() {
		if member.UserID != 0 {
			if err = dal.User.DelayUpdateUserWithBill(c, member.UserID, bill.ID); err != nil {
				return
			}
		}
	}
	return
})

func (billDalGae) UpdateBillsHolder(c context.Context, billID string, getBillsHolder dal.BillsHolderGetter) (err error) {
	return
}
