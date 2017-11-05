package models

import (
	"time"
	"github.com/strongo/decimal"
	"github.com/pquerna/ffjson/ffjson"
	"fmt"
)

//go:generate ffjson $GOFILE

type TransferReturnJson struct {
	TransferID int64
	Time time.Time
	Amount decimal.Decimal64p2
}

func (t *TransferEntity) GetReturns() (returns []TransferReturnJson) {
	if t.returns != nil && len(t.returns) == t.ReturnsCount {
		returns = make([]TransferReturnJson, t.ReturnsCount)
		copy(returns, t.returns)
		return
	}
	if len(t.ReturnsJson) == 0 {
		return
	}
	if err := ffjson.Unmarshal([]byte(t.ReturnsJson), &returns); err != nil {
		panic(err)
	}
	if len(returns) != t.ReturnsCount {
		panic(fmt.Sprintf("len(returns) != ReturnsCount: %v != %v", len(returns), t.ReturnsCount))
	}
	t.returns = make([]TransferReturnJson, len(returns))
	copy(t.returns, returns)
	return
}

func (t *TransferEntity) SetReturns(returns []TransferReturnJson) {
	t.AmountInCentsReturned = 0
	if t.ReturnsCount = len(returns); t.ReturnsCount == 0 {
		t.ReturnsJson = ""
		return
	}
	for _, r := range returns {
		t.AmountInCentsReturned += r.Amount
	}
	if data, err := ffjson.Marshal(returns); err != nil {
		panic(err)
	} else {
		t.ReturnsJson = string(data)
	}
	return
}

