package models

import (
	"fmt"
	"time"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/decimal"
)

//go:generate ffjson $GOFILE

type TransferReturnJson struct {
	TransferID int64
	Time       time.Time
	Amount     decimal.Decimal64p2 `json:",omitempty"` // TODO: For legacy records, consider removing later
}

func (t *TransferEntity) GetReturns() (returns []TransferReturnJson) {
	if t.returns != nil && len(t.returns) == t.ReturnsCount {
		returns = make([]TransferReturnJson, t.ReturnsCount)
		copy(returns, t.returns)
		return
	}
	if len(t.ReturnsJson) == 0 && len(t.ReturnTransferIDs) == 0 {
		return
	}
	if err := ffjson.Unmarshal([]byte(t.ReturnsJson), &returns); err != nil {
		panic(err)
	}
	if t.ReturnsCount == 0 {
		switch {
		case (len(t.ReturnTransferIDs) > 0 && len(returns) > 0 && len(t.ReturnTransferIDs) == len(returns)) || len(returns) > 0:
			t.ReturnsCount = len(returns)
		case len(t.ReturnTransferIDs) > 0:
			t.ReturnsCount = len(t.ReturnTransferIDs)
		default:
			panic(fmt.Sprintf("len(returns) != len(ReturnTransferIDs): %v != %v", len(returns), len(t.ReturnTransferIDs)))
		}
	}
	if len(returns) != t.ReturnsCount {
		panic(fmt.Sprintf("len(returns) != ReturnsCount: %v != %v", len(returns), t.ReturnsCount))
	}
	t.returns = make([]TransferReturnJson, len(returns))
	copy(t.returns, returns)
	return
}

func (t *TransferEntity) SetReturns(returns []TransferReturnJson) {
	if len(returns) == 0 {
		t.ReturnsCount = 0
		t.ReturnsJson = ""
		return
	}
	if data, err := ffjson.Marshal(returns); err != nil {
		panic(err)
	} else {
		t.ReturnsJson = string(data)
	}
	if len(returns) == 0 {
		t.AmountInCentsReturned = 0
	} else {
		var returnedValue decimal.Decimal64p2
		for _, r := range returns {
			returnedValue += r.Amount
		}
		if returnedValue > 0 {
			t.AmountInCentsReturned = returnedValue
		}
	}
	t.ReturnsCount = len(returns)
	t.returns = make([]TransferReturnJson, t.ReturnsCount)
	copy(t.returns, returns)
	return
}
