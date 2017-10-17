package models

import (
	"github.com/strongo/decimal"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/app/db"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/pkg/errors"
	"fmt"
	"time"
)

const BillsHistoryKind = "BillH"

type BillsHistory struct {
	db.StringID
	db.NoIntID
	*BillsHistoryEntity
}

func (BillsHistory) Kind() string {
	return BillsHistoryKind
}

func (record BillsHistory) Entity() interface{} {
	return record.BillsHistoryEntity
}

func (record *BillsHistory) SetEntity(entity interface{}) {
	if entity == nil {
		record.BillsHistoryEntity = nil
		return
	}
	record.BillsHistoryEntity = entity.(*BillsHistoryEntity)
}

var _ db.EntityHolder = (*BillsHistory)(nil)

type BillsHistoryEntity struct {
	DtCreated              time.Time
	Type                   string              `datastore:",noindex"`
	Currency               Currency            `datastore:",noindex"`
	TotalAmount            decimal.Decimal64p2 `datastore:",noindex"`
	GroupIDs               []string
	BillIDs                []string
	BillsSettlementCount             int                 `datastore:",noindex"`
	BillsSettlementJson    string              `datastore:",noindex"`
	GroupMembersJsonBefore string              `datastore:",noindex"`
	GroupMembersJsonAfter  string              `datastore:",noindex"`
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
	if entity.Type == "" {
		err = errors.New("*BillsHistoryEntity.Type is empty")
		return
	}
	if entity.BillsSettlementJson == "" {
		err = errors.New("BillsSettlementJson is empty")
		return
	}
	if entity.Currency == "" {
		err = errors.New("Currency is empty")
		return
	}
	if len(entity.GroupIDs) == 0 {
		err = errors.New("len(entity.GroupIDs) == 0")
		return
	}
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
	if entity.TotalAmount != total {
		err = fmt.Errorf("entity.TotalAmount:%v != total:%v", entity.TotalAmount, total)
		return
	}
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"BillsSettlementCount":             gaedb.IsZeroInt,
		"BillsSettlementJson":              gaedb.IsEmptyJson,
		"GroupMembersJsonBefore": gaedb.IsEmptyJson,
		"GroupMembersJsonAfter":  gaedb.IsEmptyJson,
	}); err != nil {
		return
	}
	return
}
