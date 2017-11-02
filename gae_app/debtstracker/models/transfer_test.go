package models

import (
	"github.com/strongo/decimal"
	"google.golang.org/appengine/datastore"
	"testing"
)

func TestTransfer_Save(t *testing.T) {
	saved := []struct {
		kind       string
		properties []datastore.Property
	}{}
	checkHasProperties = func(kind string, properties []datastore.Property) error {
		saved = append(saved, struct {
			kind       string
			properties []datastore.Property
		}{kind, properties})
		return nil
	}
	rub := Currency(CURRENCY_IRR)
	creator := TransferCounterpartyInfo{
		UserID:      1,
		ContactID:   2,
		ContactName: "Test1",
	}
	counterparty := TransferCounterpartyInfo{
		ContactName: "Creator 1",
	}
	transfer := NewTransferEntity(creator.UserID, false, NewAmount(rub, decimal.NewDecimal64p2FromFloat64(123.45)), &creator, &counterparty)
	if _, err := transfer.Save(); err != nil {
		t.Error(err)
	} else if len(saved) == 1 {
		if saved[0].kind != TransferKind {
			t.Errorf("saved[0].kind:'%v' != '%v'", saved[0].kind, TransferKind)
		}
	} else {
		t.Errorf("len(saved):%v != 1", len(saved))
	}
}

//func TestTransferDump(t *testing.T) {
//	now := time.Now()
//	litter.Config.HidePrivateFields = true
//	t.Log("litter.Config.HidePrivateFields = true: ", litter.Sdump(now))
//	litter.Config.HidePrivateFields = false
//	t.Log("litter.Config.HidePrivateFields = false: ", litter.Options{HidePrivateFields: false}.Sdump(now))
//}