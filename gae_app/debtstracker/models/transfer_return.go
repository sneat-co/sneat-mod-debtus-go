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

func (t *TransferEntity) AddReturn(returnTransfer TransferReturnJson) error {
	if returnTransfer.TransferID == 0 {
		return errors.New("returnTransfer.TransferID == 0")
	}
	// if returnTransfer.TransferID == t.ID {
	// 	return fmt.Errorf("returnTransfer.TransferID == t.ID => %v", t.ID)
	// }
	if returnTransfer.Time.IsZero() {
		return fmt.Errorf("returnTransfer.Time.IsZero(), ID=%v", returnTransfer.TransferID)
	}
	if returnTransfer.Amount <= 0 {
		return fmt.Errorf("transfer return amount <= 0 (ID=%v, amount=%v)", returnTransfer.TransferID, returnTransfer.Amount)
	}

	returns := t.GetReturns()
	if len(returns) != t.ReturnsCount {
		return fmt.Errorf("transfer data integrity issue: len(returns) != t.ReturnsCount => %v != %v", len(returns), t.ReturnsCount)
	}
	var returnedValue decimal.Decimal64p2
	returnTransferIDs := make([]int64, 1, len(returns) + 1)
	returnTransferIDs[0] = returnTransfer.TransferID
	for _, r := range returns {
		for _, tID := range returnTransferIDs {
			if tID == r.TransferID {
				return fmt.Errorf("duplicate return transfer ID=%v", r.TransferID)
			}
		}
		returnTransferIDs = append(returnTransferIDs, r.TransferID)
		returnedValue += r.Amount
	}
	if returnedValue != t.AmountInCentsReturned {
		return fmt.Errorf("transfer data integrity issue: sum(returns.Amount) != t.AmountInCentsReturned => %v != %v", returnedValue, t.AmountInCentsReturned)
	}

	transferAmount := t.GetAmount()
	totalDue := transferAmount.Value + t.GetInterestValue(time.Now())
	returnedValue += returnTransfer.Amount
	if returnedValue > totalDue {
		return fmt.Errorf("wrong transfer returns: returnedValue > totalDue (%v > %v)", returnedValue, totalDue)
	}
	if returnedValue == totalDue {
		t.IsOutstanding = false
	}
	returns = append(returns, returnTransfer)
	returnsJson, err := ffjson.Marshal(returns) // We'll assign to t.ReturnsJson when all checks are OK
	if err != nil {
		return errors.WithMessage(err, "failed to marshal transfer returns")
	}
	t.ReturnsJson = string(returnsJson)
	t.AmountInCentsReturned = returnedValue
	t.ReturnsCount = len(returns)
	t.returns = returns
	return nil
}
