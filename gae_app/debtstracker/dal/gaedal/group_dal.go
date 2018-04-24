package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
)

var _ dal.GroupDal = (*GroupDalGae)(nil)

func NewGroupKey(c context.Context, groupID string) *datastore.Key {
	if groupID == "" {
		panic("groupID is empty string")
	}
	return gaedb.NewKey(c, models.GroupKind, groupID, 0, nil)
}

type GroupDalGae struct {
}

func NewGroupDalGae() GroupDalGae {
	return GroupDalGae{}
}

func (_ GroupDalGae) InsertGroup(c context.Context, groupEntity *models.GroupEntity) (group models.Group, err error) {
	group.GroupEntity = groupEntity
	err = dal.InsertWithRandomStringID(c, &group, models.GroupIdLen)
	return
}

func (_ GroupDalGae) SaveGroup(c context.Context, group models.Group) (err error) {
	if _, err = gaedb.Put(c, NewGroupKey(c, group.ID), group.GroupEntity); err != nil {
		return
	}
	return
}

func (GroupDalGae) GetGroupByID(c context.Context, groupID string) (group models.Group, err error) {
	if group.ID = groupID; group.ID == "" {
		panic("groupID is empty string")
	}
	group.ID = groupID
	if err = dal.DB.Get(c, &group); err != nil {
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
	bill, err := dal.Bill.GetBillByID(c, billID)
	if err != nil {
		return
	}
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var group models.Group
		if group, err = dal.Group.GetGroupByID(c, groupID); err != nil {
			return err
		}
		var changed bool
		if changed, err = group.AddBill(bill); err != nil {
			return err
		} else if changed {
			if err = dal.Group.SaveGroup(c, group); err != nil {
				return err
			}
		}
		return
	}, db.SingleGroupTransaction); err != nil {
		return
	}
	return
})
