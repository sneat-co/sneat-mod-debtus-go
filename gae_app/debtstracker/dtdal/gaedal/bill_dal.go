package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/app/gae"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/delay"
)

type billDalGae struct {
}

// var _ dtdal.BillDal = (*billDalGae)(nil) // Make sure we implement interface

func newBillDalGae() billDalGae {
	return billDalGae{}
}

func (billDalGae) SaveBill(c context.Context, bill models.Bill) (err error) {
	if err = dtdal.DB.Update(c, &bill); err != nil {
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
	if bill, err = facade.GetBillByID(c, billID); err != nil {
		if dal.IsNotFound(err) {
			log.Warningf(c, err.Error())
			err = nil
		}
		return
	}
	if userGroupID := bill.GetUserGroupID(); userGroupID != "" {
		if err = dtdal.Group.DelayUpdateGroupWithBill(c, userGroupID, bill.ID); err != nil {
			return
		}
	}
	for _, member := range bill.GetBillMembers() {
		if member.UserID != "" {
			if err = dtdal.User.DelayUpdateUserWithBill(c, member.UserID, bill.ID); err != nil {
				return
			}
		}
	}
	return
})

func (billDalGae) UpdateBillsHolder(c context.Context, billID string, getBillsHolder dtdal.BillsHolderGetter) (err error) {
	return
}
