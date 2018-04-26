package gaedal

import (
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"context"
	"google.golang.org/appengine/delay"
)

type billDalGae struct {
}

var _ dal.BillDal = (*billDalGae)(nil) // Make sure we implement interface

func newBillDalGae() billDalGae {
	return billDalGae{}
}

func (billDalGae) GetBillByID(c context.Context, billID string) (bill models.Bill, err error) {
	if billID == "" {
		panic("billID is empty string")
	}
	bill.ID = billID
	bill.BillEntity = new(models.BillEntity)
	if err = dal.DB.Get(c, &bill); err != nil {
		bill.ID = ""
		bill.BillEntity = nil
	}
	return
}

func (billDalGae) GetBillsByIDs(c context.Context, billIDs []string) (bills []models.Bill, err error) {
	entityHolders := make([]db.EntityHolder, len(billIDs))
	for i, id := range billIDs {
		entityHolders[i] = &models.Bill{StringID: db.StringID{ID: id}, BillEntity: new(models.BillEntity)}
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

func (billDalGae) InsertBillEntity(c context.Context, billEntity *models.BillEntity) (bill models.Bill, err error) {
	if billEntity == nil {
		panic("billEntity == nil")
	}
	if billEntity.CreatorUserID == "" {
		panic("CreatorUserID == 0")
	}
	if billEntity.AmountTotal == 0 {
		panic("AmountTotal == 0")
	}

	billEntity.DtCreated = time.Now()
	bill.BillEntity = billEntity

	err = dal.InsertWithRandomStringID(c, &bill, models.BillIdLen)
	return
}

func (billDalGae) SaveBill(c context.Context, bill models.Bill) (err error) {
	if err = dal.DB.Update(c, &bill); err != nil {
		return
	}
	if err = DelayUpdateUsersWithBill(c, bill.ID, bill.UserIDs); err != nil {
		return
	}
	return
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
	if userGroupID := bill.UserGroupID(); userGroupID != "" {
		if err = dal.Group.DelayUpdateGroupWithBill(c, userGroupID, bill.ID); err != nil {
			return
		}
	}
	for _, member := range bill.GetBillMembers() {
		if member.UserID != "" {
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
