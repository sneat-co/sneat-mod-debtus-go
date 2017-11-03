package models

import (
	"time"
	"github.com/strongo/decimal"
	"github.com/pquerna/ffjson/ffjson"
)

//go:generate ffjson $GOFILE

type TransferReturnJson struct {
	TransferID int64
	Time time.Time
	Amount decimal.Decimal64p2
}

func (t *TransferEntity) GetReturns() (returns []TransferReturnJson) {
	if len(t.ReturnsJson) == 0 {
		return
	}
	if err := ffjson.Unmarshal([]byte(t.ReturnsJson), &returns); err != nil {
		panic(err)
	}
	return
}

func (t *TransferEntity) SetReturns(returns []TransferReturnJson) {
	if len(returns) == 0 {
		t.ReturnsJson = ""
	}
	if data, err := ffjson.Marshal(returns); err != nil {
		panic(err)
	} else {
		t.ReturnsJson = string(data)
	}
	return
}

