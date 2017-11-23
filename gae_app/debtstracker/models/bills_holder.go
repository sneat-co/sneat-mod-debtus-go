package models

import (
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/db/gaedb"
)

type billsHolder struct {
	OutstandingBillsCount int    `datastore:",noindex,omitempty"`
	OutstandingBillsJson  string `datastore:",noindex,omitempty"`
}

func (entity *billsHolder) GetOutstandingBills() (outstandingBills []BillJson, err error) {
	if entity.OutstandingBillsJson == "" {
		return
	}
	if err = ffjson.Unmarshal([]byte(entity.OutstandingBillsJson), &outstandingBills); err != nil {
		return
	}
	if entity.OutstandingBillsCount != len(outstandingBills) {
		err = errors.WithMessage(ErrJsonCountMismatch, "len([]BillJson) != OutstandingBillsCount")
	}
	return
}

func (entity *billsHolder) SetOutstandingBills(outstandingBills []BillJson) (changed bool, err error) {
	var data []byte
	if data, err = ffjson.Marshal(outstandingBills); err != nil {
		return
	}
	json := string(data)
	if json == "[]" {
		json = ""
	}
	entity.OutstandingBillsCount = len(outstandingBills)
	if changed = json != entity.OutstandingBillsJson; changed {
		entity.OutstandingBillsJson = json
	}
	return
}

func init() {
	userPropertiesToClean["OutstandingBillsJson"] = gaedb.IsEmptyJson
	groupPropertiesToClean["OutstandingBillsJson"] = gaedb.IsEmptyJson
}