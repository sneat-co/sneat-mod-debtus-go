package models

//go:generate ffjson $GOFILE

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"time"
)

type LastTransfer struct {
	ID int64
	At time.Time
}

type UserCounterpartyTransfersInfo struct {
	Count int
	Last  LastTransfer
}

type UserContactJson struct {
	ID          int64
	Name        string
	Status      string                         `json:",omitempty"`
	UserID      int64                          `json:",omitempty"` // TODO: new prop, update in map reduce and change code!
	TgUserID    int64                          `json:",omitempty"`
	BalanceJson *json.RawMessage               `json:"Balance,omitempty"`
	Transfers   *UserCounterpartyTransfersInfo `json:",omitempty"`
}

func (o UserContactJson) Balance() (balance Balance, err error) {
	balance = make(Balance)
	if o.BalanceJson != nil {
		if err = ffjson.Unmarshal(*o.BalanceJson, &balance); err != nil { // TODO: Migrate to ffjson.UnmarshalFast() ?
			err = errors.Wrapf(err, "Failed to unmarshal BalanceJson for counterparty with ID=%v", o.ID)
			return
		}
	}
	return
}

func NewUserCountactJson(counterpartyID int64, status, name string, balanced Balanced) UserContactJson {
	result := UserContactJson{
		ID:     counterpartyID,
		Status: status,
		Name:   name,
	}
	if balanced.BalanceJson != "" {
		balance := json.RawMessage(balanced.BalanceJson)
		result.BalanceJson = &balance
	}
	if balanced.CountOfTransfers != 0 {
		if balanced.LastTransferID == 0 {
			panic(fmt.Sprintf("balanced.CountOfTransfers:%v != 0 && balanced.LastTransferID == 0", balanced.CountOfTransfers))
		}
		if balanced.LastTransferAt.IsZero() {
			panic(fmt.Sprintf("balanced.CountOfTransfers:%v != 0 && balanced.LastTransferAt.IsZero():true", balanced.CountOfTransfers))
		}
		result.Transfers = &UserCounterpartyTransfersInfo{
			Count: balanced.CountOfTransfers,
			Last: LastTransfer{
				ID: balanced.LastTransferID,
				At: balanced.LastTransferAt,
			},
		}
	}
	return result
}
