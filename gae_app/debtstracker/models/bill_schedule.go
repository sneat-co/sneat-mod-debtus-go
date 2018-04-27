package models

import (
	"time"

	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

type BillScheduleStatus string

const (
	BillScheduleStatusDraft    BillScheduleStatus = "draft"
	BillScheduleStatusActive   BillScheduleStatus = STATUS_ACTIVE
	BillScheduleStatusArchived BillScheduleStatus = STATUS_ARCHIVED
	//BillScheduleStatusDeleted  BillScheduleStatus = STATUS_DELETED
)

type Period string

const (
	PeriodWeekly  Period = "weekly"
	PeriodMonthly Period = "monthly"
	PeriodYearly  Period = "yearly"
)

const BillScheduleKind = "BillSchedule"

type BillSchedule struct {
	db.IntegerID
	*BillScheduleEntity
}

type BillScheduleEntity struct {
	BillCommon
	/* Repeat examples (RepeatPeriod:RepeatOn)
	* weekly:monday
	* monthly:2 - 2nd day of each month. possible values 1-28
	// * monthly:first-monday
	// * yearly:1-jan ???
	*/
	BillsCount        int    `datastore:",noindex"`
	CreatedFromBillID string `datastore:",noindex"`
	RepeatPeriod      Period `datastore:",noindex"`
	RepeatOn          string `datastore:",noindex"`
	IsAutoTransfer    bool   `datastore:",noindex"`

	LastBillID string    `datastore:",noindex"`
	DtLast     time.Time `datastore:",noindex"`
	DtNext     time.Time
}

func (entity *BillScheduleEntity) Load(ps []datastore.Property) error {
	ps = entity.BillCommon.load(ps)
	return datastore.LoadStruct(entity, ps)
}

func (entity *BillScheduleEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = entity.BillCommon.save(properties); err != nil {
		return
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"DtLast":            gaedb.IsZeroTime,
		"DtNext":            gaedb.IsZeroTime,
		"LastBillID":        gaedb.IsZeroInt,
		"IsAutoTransfer":    gaedb.IsZeroBool,
		"BillsCount":        gaedb.IsZeroInt,
		"CreatedFromBillID": gaedb.IsZeroInt,
	}); err != nil {
		return
	}
	return
}

func (BillSchedule) Kind() string {
	return BillKind
}

func (bill BillSchedule) IntID() int64 {
	return bill.ID
}

func (bill *BillSchedule) Entity() interface{} {
	if bill.BillScheduleEntity == nil {
		bill.BillScheduleEntity = new(BillScheduleEntity)
	}
	return bill.BillScheduleEntity
}

func (bill *BillSchedule) SetEntity(entity interface{}) {
	bill.BillScheduleEntity = entity.(*BillScheduleEntity)
}
