package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/delay"
	"github.com/strongo/app/gae"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
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
	dal.InsertWithRandomStringID(c, &group, 8)
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
	return group, dal.DB.Get(c, &group)
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
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		group, err := dal.Group.GetGroupByID(c, groupID)
		if err != nil {
			return err
		}
		outstandingBills, err := group.GetOutstandingBills()
		if err != nil {
			return err
		}
		changed := false

		for i, b := range outstandingBills {
			if b.ID == billID {
				if b.Name != bill.Name {
					outstandingBills[i].Name = bill.Name
					changed = true
				}
				if b.MembersCount != bill.MembersCount {
					outstandingBills[i].MembersCount = bill.MembersCount
					changed = true
				}
				if b.Total != bill.AmountTotal {
					outstandingBills[i].Total = bill.AmountTotal
					changed = true
				}
				goto updated
			}
		}
		outstandingBills = append(outstandingBills, models.BillJson{
			ID: bill.ID,
			Name: bill.Name,
			MembersCount: bill.MembersCount,
			Total: bill.AmountTotal,
			Currency: bill.Currency,
		})
		changed = true
	updated:
		if changed {
			if changed, err = group.SetOutstandingBills(outstandingBills); err != nil {
				return err
			} else if changed {
				if err = dal.Group.SaveGroup(c, group); err != nil {
					return err
				}
			}
		}
		return nil
	}, db.SingleGroupTransaction)
	if err != nil {
		return
	}
	return
})
