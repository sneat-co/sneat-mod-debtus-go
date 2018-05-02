package models

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
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

func (t *TransferEntity) SetReturns(returns []TransferReturnJson) error {
	if len(returns) == 0 {
		t.AmountInCentsReturned = 0
		t.ReturnsCount = 0
		t.ReturnsJson = ""
		return nil
	}

	transferAmount := t.GetAmount()
	totalDue := transferAmount.Value + t.GetInterestValue(time.Now())

	var returnedValue decimal.Decimal64p2
	returnTransferIDs := make([]int64, 0, len(returns))
	for i, r := range returns {
		if r.TransferID == 0 {
			return fmt.Errorf("return is missing TransferID (i=%d)", i)
		}
		if r.Amount <= 0 {
			return fmt.Errorf("transfer return amount <= 0 (ID=%v, amount=%v)", r.TransferID, r.Amount)
		}
		for _, tID := range returnTransferIDs {
			if tID == r.TransferID {
				return fmt.Errorf("duplicate return transfer ID=%v", r.TransferID)
			}
		}
		returnTransferIDs = append(returnTransferIDs, r.TransferID)
		returnedValue += r.Amount
	}
	if returnedValue > totalDue {
		return fmt.Errorf("wrong transfer returns: returnedValue > totalDue (%v > %v)", returnedValue, totalDue)
	}
	if returnedValue == totalDue {
		t.IsOutstanding = false
	}
	returnsJson, err := ffjson.Marshal(returns) // We'll assign to t.ReturnsJson when all checks are OK
	if err != nil {
		return errors.WithMessage(err, "failed to marshal transfer returns")
	}
	t.AmountInCentsReturned = returnedValue
	t.ReturnsJson = string(returnsJson)
	t.ReturnsCount = len(returns)
	t.returns = make([]TransferReturnJson, t.ReturnsCount)
	copy(t.returns, returns)

	return nil
}
