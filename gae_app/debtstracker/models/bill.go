package models

import (
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/datastore"
	"time"
	"fmt"
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
	//BalanceJson      string    `datastore:",noindex"`
	//BalanceVersion   int       `datastore:",noindex"`
	//balanceVersion   int       `datastore:"-"`
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
	if entity == nil {
		bill.BillEntity = nil
	} else {
		bill.BillEntity = entity.(*BillEntity)
	}
}

func (entity *BillEntity) Load(ps []datastore.Property) error {
	if err := entity.BillCommon.load(ps); err != nil {
		return err
	}
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

func (entity *BillEntity) GetBalance() (BillBalanceByCurrencyAndMember){
	members := entity.GetBillMembers()
	totalsByMember := make(BillBalanceByMember, len(members))

	for i, member := range members {

		if member.Owes < 0 {
			panic(fmt.Sprintf("member[%d].Owes < 0: %v", i, member.Owes))
		} else if member.Paid < 0 {
			panic(fmt.Sprintf("member[%d].Paid < 0: %v", i, member.Paid))
		}

		if member.Owes != 0 || member.Paid != 0 {
			totalsByMember[member.ID] = BillMemberBalance{
				Owes: member.Owes,
				Paid: member.Paid,
			}
		}
	}
	return BillBalanceByCurrencyAndMember{entity.Currency: totalsByMember}
}
