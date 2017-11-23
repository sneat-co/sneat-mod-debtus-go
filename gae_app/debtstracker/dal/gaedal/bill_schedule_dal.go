package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/db/gaedb"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

var _ dal.BillScheduleDal = (*BillScheduleDalGae)(nil)

type BillScheduleDalGae struct {
}

func NewBillScheduleDalGae() BillScheduleDalGae {
	return BillScheduleDalGae{}
}

func NewBillScheduleKey(c context.Context, id int64) *datastore.Key {
	return datastore.NewKey(c, models.BillScheduleKind, "", id, nil)
}

func NewBillScheduleIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.BillScheduleKind, nil)
}

func (BillScheduleDalGae) GetBillScheduleByID(c context.Context, id int64) (billSchedule models.BillSchedule, err error) {
	billSchedule.BillScheduleEntity = new(models.BillScheduleEntity)
	if err = gaedb.Get(c, NewBillScheduleKey(c, id), billSchedule.BillScheduleEntity); err != nil {
		billSchedule.BillScheduleEntity = nil
		return
	}
	return
}

func (BillScheduleDalGae) InsertBillSchedule(c context.Context, billScheduleEntity *models.BillScheduleEntity) (billSchedule models.BillSchedule, err error) {
	key := NewBillScheduleIncompleteKey(c)
	if key, err = gaedb.Put(c, key, billScheduleEntity); err != nil {
		return
	}
	billSchedule.ID = key.IntID()
	return
}

func (BillScheduleDalGae) UpdateBillSchedule(c context.Context, billSchedule models.BillSchedule) (err error) {
	_, err = gaedb.Put(c, NewBillScheduleKey(c, billSchedule.ID), billSchedule.BillScheduleEntity)
	return
}
