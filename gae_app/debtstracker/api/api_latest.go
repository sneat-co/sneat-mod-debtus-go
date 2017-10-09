package api

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"net/http"
	"sync"
)

func handleAdminLatestUsers(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	users, err := dal.Admin.LatestUsers(c)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	var buffer bytes.Buffer
	buffer.WriteString("[")
	lastIndex := len(users) - 1
	var wg sync.WaitGroup
	type CounterpartyDto struct {
		Id      int64
		UserID  int64 `json:",omitempty"`
		Name    string
		Balance *json.RawMessage `json:",omitempty"`
	}
	type Record struct {
		Id                     int64
		Name                   string
		Counterparties         []CounterpartyDto
		Transfers              int
		CountOfReceiptsCreated int
		InvitedByUser          *struct {
			Id   int64
			Name string
		} `json:",omitempty"`
		//InvitedByUserID int64 `json:",omitempty"`
		//InvitedByUserName string `json:",omitempty"`
		Balance         *json.RawMessage `json:",omitempty"`
		TelegramUserIDs []int64          `json:",omitempty"`
	}
	records := make([]*Record, len(users))
	for i, user := range users {
		records[i] = &Record{
			Id:                     user.ID,
			Name:                   user.FullName(),
			Transfers:              user.CountOfTransfers,
			CountOfReceiptsCreated: user.CountOfReceiptsCreated,
			TelegramUserIDs:        user.GetTelegramUserIDs(),
		}
		if user.BalanceJson != "" {
			balance := json.RawMessage(user.BalanceJson)
			records[i].Balance = &balance
		}
		userCounterpartiesIDs := user.CounterpartiesIDs()
		if len(userCounterpartiesIDs) > 0 {
			wg.Add(1)
			go func(i int, userCounterpartiesIDs []int64) {
				counterparties, err := dal.Contact.GetContactsByIDs(c, userCounterpartiesIDs)
				if err != nil {
					log.Errorf(c, errors.Wrapf(err, "Failed to get counterparties by ids=%v", userCounterpartiesIDs).Error())
					wg.Done()
					return
				}
				record := records[i]
				for j, counterparty := range counterparties {
					counterpartyDto := CounterpartyDto{
						Id:     userCounterpartiesIDs[j],
						UserID: counterparty.CounterpartyUserID,
						Name:   counterparty.FullName(),
					}
					if counterparty.BalanceJson != "" {
						balance := json.RawMessage(counterparty.BalanceJson)
						counterpartyDto.Balance = &balance
					}
					record.Counterparties = append(record.Counterparties, counterpartyDto)
				}
				log.Debugf(c, "Contacts goroutine completed.")
				wg.Done()
			}(i, userCounterpartiesIDs)
		}
		if user.InvitedByUserID != 0 {
			wg.Add(1)
			go func(i int, userID int64) {
				inviter, err := dal.User.GetUserByID(c, userID)
				if err != nil {
					log.Errorf(c, errors.Wrapf(err, "Failed to get user by id=%v", userID).Error())
					return
				}
				records[i].InvitedByUser = &struct {
					Id   int64
					Name string
				}{
					userID,
					inviter.FullName(),
				}
				log.Debugf(c, "User goroutine completed.")
				wg.Done()
			}(i, user.InvitedByUserID)
		}
	}

	wg.Wait()

	for i, record := range records {
		if userBytes, err := json.Marshal(record); err != nil {
			log.Errorf(c, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			buffer.Write(userBytes)
		}
		if i < lastIndex {
			buffer.Write([]byte(","))
		}
	}

	buffer.WriteString("]")
	header := w.Header()
	header.Add("Content-Type", "application/json")
	header.Add("Access-Control-Allow-Origin", "*")
	if _, err = w.Write(buffer.Bytes()); err != nil {
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}
