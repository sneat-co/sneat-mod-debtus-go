package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/app/gae"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/delay"
)

var _ dtdal.GroupDal = (*GroupDalGae)(nil)

type GroupDalGae struct { // TODO: Obsolete naming with migration to Dalgo
}

func NewGroupDalGae() GroupDalGae {
	return GroupDalGae{}
}

func (GroupDalGae) InsertGroup(c context.Context, tx dal.ReadwriteTransaction, groupEntity *models.GroupEntity) (group models.Group, err error) {
	group = models.NewGroup("", groupEntity)
	err = dtdal.InsertWithRandomStringID(c, tx, group.Record)
	return
}

func (GroupDalGae) SaveGroup(c context.Context, tx dal.ReadwriteTransaction, group models.Group) (err error) {
	if err = tx.Set(c, group.Record); err != nil {
		return
	}
	return
}

func (GroupDalGae) GetGroupByID(c context.Context, tx dal.ReadSession, groupID string) (group models.Group, err error) {
	if group.ID = groupID; group.ID == "" {
		panic("groupID is empty string")
	}
	group = models.NewGroup(groupID, nil)
	if err = tx.Get(c, group.Record); err != nil {
		return
	}
	return
}

func (GroupDalGae) DelayUpdateGroupWithBill(c context.Context, groupID, billID string) (err error) {
	if err = gae.CallDelayFunc(c, common.QUEUE_BILLS, "UpdateGroupWithBill", delayedUpdateGroupWithBill, groupID, billID); err != nil {
		return
	}
	return
}

var delayedUpdateGroupWithBill = delay.Func("delayedUpdateWithBill", func(c context.Context, groupID, billID string) (err error) {
	log.Debugf(c, "delayedUpdateGroupWithBill(groupID=%d, billID=%d)", groupID, billID)
	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}
	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
		bill, err := facade.GetBillByID(c, tx, billID)
		if err != nil {
			return
		}
		var group models.Group
		if group, err = dtdal.Group.GetGroupByID(c, tx, groupID); err != nil {
			return err
		}
		var changed bool
		if changed, err = group.Data.AddBill(bill); err != nil {
			return err
		} else if changed {
			if err = dtdal.Group.SaveGroup(c, tx, group); err != nil {
				return err
			}
		}
		return
	}); err != nil {
		return
	}
	return
})
