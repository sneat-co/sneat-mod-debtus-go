package models

import (
	"time"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/app/db"
)

const (
	BillKind = "Bill"
)

const (
	BillStatusDraft   = "draft"
	BillStatusActive  = "active"
	BillStatusSettled = "settled"
)

var (
	BillStatuses = [3]string{
		BillStatusDraft,
		BillStatusActive,
		BillStatusSettled,
	}
	BillSplitModes = [5]SplitMode{
		SplitModeAdjustment,
		SplitModeEqually,
		SplitModeExactAmount,
		SplitModePercentage,
		SplitModeShare,
	}
)

func IsValidBillSplit(split SplitMode) bool {
	for _, v := range BillSplitModes {
		if split == v {
			return true
		}
	}
	return false
}

func IsValidBillStatus(status string) bool {
	for _, v := range BillStatuses {
		if status == v {
			return true
		}
	}
	return false
}

type BillEntity struct {
	BillCommon
	DtDueToPay       time.Time `datastore:",noindex"` // TODO: Document diff between DtDueToPay & DtDueToCollect
	DtDueToCollect   time.Time `datastore:",noindex"`
	LocaleByMessage  []string  `datastore:",noindex"`
	TgChatMessageIDs []string  `datastore:",noindex"`
	SplitID          int64     `datastore:",noindex"`
}

func NewBillEntity(data BillCommon) *BillEntity {
	return &BillEntity{
		BillCommon: data,
	}
}

type Bill struct {
	db.StringID
	db.NoIntID
	*BillEntity
}

var _ db.EntityHolder = (*Bill)(nil)

func (Bill) Kind() string {
	return BillKind
}

func (bill *Bill) Entity() interface{} {
	if bill.BillEntity == nil {
		bill.BillEntity = new(BillEntity)
	}
	return bill.BillEntity
}

func (bill *Bill) SetEntity(entity interface{}) {
	bill.BillEntity = entity.(*BillEntity)
}

func (entity *BillEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *BillEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = entity.BillCommon.save(properties); err != nil {
		return
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"DtDueToPay":     gaedb.IsZeroTime,
		"DtDueToCollect": gaedb.IsZeroTime,
	}); err != nil {
		return
	}
	return
}
