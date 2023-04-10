package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
)

var _ dtdal.BillScheduleDal = (*BillScheduleDalGae)(nil)

type BillScheduleDalGae struct {
}

func NewBillScheduleDalGae() BillScheduleDalGae {
	return BillScheduleDalGae{}
}

func NewBillScheduleKey(id int64) *dal.Key {
	return dal.NewKeyWithID(models.BillScheduleKind, id)
}

func NewBillScheduleIncompleteKey() *dal.Key {
	return dal.NewKey(models.BillScheduleKind)
}

func (BillScheduleDalGae) GetBillScheduleByID(c context.Context, id int64) (models.BillSchedule, error) {
	key := NewBillScheduleKey(id)
	data := new(models.BillScheduleEntity)
	billSchedule := models.BillSchedule{
		WithID: record.WithID[int64]{
			ID:     id,
			Key:    key,
			Record: dal.NewRecordWithData(key, data),
		},
		Data: data,
	}
	db, err := GetDatabase(c)
	if err != nil {
		return billSchedule, err
	}
	if err = db.Get(c, billSchedule.Record); err != nil {
		return billSchedule, err
	}
	return billSchedule, err
}

func (BillScheduleDalGae) InsertBillSchedule(c context.Context, billScheduleEntity *models.BillScheduleEntity) (billSchedule models.BillSchedule, err error) {
	_ = NewBillScheduleIncompleteKey()
	panic("TODO: implement me")
	//key := NewBillScheduleIncompleteKey()
	//if key, err = gaedb.Put(c, key, billScheduleEntity); err != nil {
	//	return
	//}
	//billSchedule.ID = key.ID.(int)
	//return
}

func (BillScheduleDalGae) UpdateBillSchedule(c context.Context, billSchedule models.BillSchedule) error {
	db, err := GetDatabase(c)
	if err != nil {
		return err
	}
	return db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(c, billSchedule.Record)
	})
}
