package models

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/decimal"
	"google.golang.org/appengine/datastore"
)

const (
	BillsHistoryKind = "BillH"
)

type BillsHistory struct {
	db.StringID
	*BillsHistoryEntity
}

func (BillsHistory) Kind() string {
	return BillsHistoryKind
}

func (record BillsHistory) Entity() interface{} {
	return record.BillsHistoryEntity
}

func (BillsHistory) NewEntity() interface{} {
	return new(BillsHistoryEntity)
}

func (record *BillsHistory) SetEntity(entity interface{}) {
	if entity == nil {
		record.BillsHistoryEntity = nil
	} else {
		record.BillsHistoryEntity = entity.(*BillsHistoryEntity)
	}
}

var _ db.EntityHolder = (*BillsHistory)(nil)

type BillsHistoryEntity struct {
	DtCreated              time.Time
	UserID                 string
	StatusOld              string              `datastore:",noindex"`
	StatusNew              string              `datastore:",noindex"`
	Action                 BillHistoryAction   `datastore:",noindex"`
	Currency               Currency            `datastore:",noindex"`
	TotalAmountDiff        decimal.Decimal64p2 `datastore:",noindex"`
	TotalAmountBefore      decimal.Decimal64p2 `datastore:",noindex"`
	TotalAmountAfter       decimal.Decimal64p2 `datastore:",noindex"`
	GroupIDs               []string
	BillIDs                []string
	BillsSettlementCount   int    `datastore:",noindex"`
	BillsSettlementJson    string `datastore:",noindex"`
	GroupMembersJsonBefore string `datastore:",noindex"`
	GroupMembersJsonAfter  string `datastore:",noindex"`
}

func (entity *BillsHistoryEntity) BillSettlements() (billSettlements []BillSettlementJson) {
	billSettlements = make([]BillSettlementJson, 0, entity.BillsSettlementCount)
	if err := ffjson.Unmarshal([]byte(entity.BillsSettlementJson), &billSettlements); err != nil {
		panic(err)
	}
	return
}

func (entity *BillsHistoryEntity) SetBillSettlements(groupID string, billSettlements []BillSettlementJson) { // TODO: Enable support for multiple groups
	if data, err := ffjson.Marshal(&billSettlements); err != nil {
		panic(err)
	} else {
		entity.BillsSettlementJson = string(data)
		entity.BillsSettlementCount = len(billSettlements)
		entity.BillIDs = make([]string, len(billSettlements))
		entity.GroupIDs = make([]string, 0, 1)
		for i, b := range billSettlements {
			entity.BillIDs[i] = b.BillID
			if b.GroupID != "" {
				for _, groupID := range entity.GroupIDs {
					if groupID == b.GroupID {
						goto groupFound
					}
				}
				entity.GroupIDs = append(entity.GroupIDs, b.GroupID)
			groupFound:
			}
		}
	}
}

func (entity *BillsHistoryEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *BillsHistoryEntity) Save() (properties []datastore.Property, err error) {
	if entity.DtCreated.IsZero() {
		entity.DtCreated = time.Now()
	}
	if entity.Action == "" {
		err = errors.New("*BillsHistoryEntity.Action is empty")
		return
	}
	if entity.Action == BillHistoryActionSettled && entity.BillsSettlementJson == "" {
		err = errors.New("BillsSettlementJson is empty")
		return
	}
	//if entity.Currency == "" {
	//	err = errors.New("Currency is empty")
	//	return
	//}
	if len(entity.GroupIDs) == 0 {
		err = errors.New("len(entity.GroupIDs) == 0")
		return
	}
	if entity.BillsSettlementJson == "" {
		if entity.BillsSettlementCount != 0 {
			err = errors.New("BillsSettlementJson is not empty && BillsSettlementCount !=  0")
			return
		}
	} else {
		bills := entity.BillSettlements()
		if entity.BillsSettlementCount != len(bills) {
			err = errors.New("BillsCount != len(bills)")
			return
		}
		var total decimal.Decimal64p2
		for i, b := range bills {
			total += b.Amount
			if entity.BillIDs[i] != b.BillID {
				err = fmt.Errorf("entity.BillIDs[%d]:%v != b.BillID:%v", i, entity.BillIDs[i], b.BillID)
			}
		}
		if entity.TotalAmountAfter != total {
			err = fmt.Errorf("entity.TotalAmount:%v != total:%v", entity.TotalAmountAfter, total)
			return
		}
	}
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"Currency":               gaedb.IsEmptyString,
		"TotalAmountDiff":        gaedb.IsZeroInt,
		"TotalAmountBefore":      gaedb.IsZeroInt,
		"TotalAmountAfter":       gaedb.IsZeroInt,
		"BillsSettlementCount":   gaedb.IsZeroInt,
		"BillsSettlementJson":    gaedb.IsEmptyJSON,
		"GroupMembersJsonBefore": gaedb.IsEmptyJSON,
		"GroupMembersJsonAfter":  gaedb.IsEmptyJSON,
	}); err != nil {
		return
	}
	return
}

type BillHistoryAction string

const (
	BillHistoryActionCreated     BillHistoryAction = "created"
	BillHistoryActionDeleted     BillHistoryAction = "deleted"
	BillHistoryActionRestored    BillHistoryAction = "restored"
	BillHistoryActionMemberAdded BillHistoryAction = "member-added"
	BillHistoryActionSettled     BillHistoryAction = "settled"
)

func NewBillHistoryBillCreated(bill Bill, groupEntity *GroupEntity) (record BillsHistory) {
	record = BillsHistory{
		BillsHistoryEntity: &BillsHistoryEntity{
			Currency:         bill.Currency,
			UserID:           bill.CreatorUserID,
			TotalAmountAfter: bill.AmountTotal,
			Action:           BillHistoryActionCreated,
			BillIDs:          []string{bill.ID},
			GroupIDs:         []string{bill.userGroupID},
		},
	}
	if groupEntity != nil {
		record.GroupMembersJsonAfter = groupEntity.MembersJson
	}
	return
}

func NewBillHistoryMemberAdded(userID string, bill Bill, totalAboutBefore decimal.Decimal64p2, groupMemberJsonBefore, groupMemberJsonAfter string) (record BillsHistory) {
	record = BillsHistory{
		BillsHistoryEntity: &BillsHistoryEntity{
			UserID:            userID,
			Currency:          bill.Currency,
			TotalAmountBefore: totalAboutBefore,
			TotalAmountAfter:  bill.AmountTotal,
			Action:            BillHistoryActionMemberAdded,
			BillIDs:           []string{bill.ID},
			GroupIDs:          []string{bill.userGroupID},
		},
	}
	record.GroupMembersJsonBefore = groupMemberJsonBefore
	record.GroupMembersJsonAfter = groupMemberJsonAfter
	return
}

func NewBillHistoryBillDeleted(userID string, bill Bill) (record BillsHistory) {
	return BillsHistory{
		BillsHistoryEntity: &BillsHistoryEntity{
			StatusOld:         bill.Status,
			StatusNew:         BillStatusDeleted,
			UserID:            userID,
			Currency:          bill.Currency,
			TotalAmountBefore: bill.AmountTotal,
			TotalAmountAfter:  bill.AmountTotal,
			Action:            BillHistoryActionMemberAdded,
			BillIDs:           []string{bill.ID},
			GroupIDs:          []string{bill.userGroupID},
		},
	}
}

func NewBillHistoryBillRestored(userID string, bill Bill) (record BillsHistory) {
	return BillsHistory{
		BillsHistoryEntity: &BillsHistoryEntity{
			StatusOld:         BillStatusDeleted,
			StatusNew:         bill.Status,
			UserID:            userID,
			Currency:          bill.Currency,
			TotalAmountBefore: bill.AmountTotal,
			TotalAmountAfter:  bill.AmountTotal,
			Action:            BillHistoryActionMemberAdded,
			BillIDs:           []string{bill.ID},
			GroupIDs:          []string{bill.userGroupID},
		},
	}
}
