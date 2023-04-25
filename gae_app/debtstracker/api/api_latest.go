package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"context"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/api/dto"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/auth"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"github.com/strongo/log"
)

func handleAdminLatestUsers(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	users, err := dtdal.Admin.LatestUsers(c)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	var buffer bytes.Buffer
	buffer.WriteString("[")
	lastIndex := len(users) - 1
	var wg sync.WaitGroup
	records := make([]*dto.Record, len(users))
	for i, user := range users {
		records[i] = &dto.Record{
			Id:                     user.ID,
			Name:                   user.Data.FullName(),
			Transfers:              user.Data.CountOfTransfers,
			CountOfReceiptsCreated: user.Data.CountOfReceiptsCreated,
			TelegramUserIDs:        user.Data.GetTelegramUserIDs(),
		}
		if user.Data.BalanceJson != "" {
			balance := json.RawMessage(user.Data.BalanceJson)
			records[i].Balance = &balance
		}
		userCounterpartiesIDs := user.Data.ContactIDs()
		if len(userCounterpartiesIDs) > 0 {
			wg.Add(1)
			go func(i int, userCounterpartiesIDs []int64) {
				counterparties, err := facade.GetContactsByIDs(c, nil, userCounterpartiesIDs)
				if err != nil {
					log.Errorf(c, fmt.Errorf("failed to get counterparties by ids=%+v: %w", userCounterpartiesIDs, err).Error())
					wg.Done()
					return
				}
				record := records[i]
				for j, counterparty := range counterparties {
					counterpartyDto := dto.CounterpartyDto{
						Id:     userCounterpartiesIDs[j],
						UserID: counterparty.Data.CounterpartyUserID,
						Name:   counterparty.Data.FullName(),
					}
					if counterparty.Data.BalanceJson != "" {
						balance := json.RawMessage(counterparty.Data.BalanceJson)
						counterpartyDto.Balance = &balance
					}
					record.Counterparties = append(record.Counterparties, counterpartyDto)
				}
				log.Debugf(c, "Contacts goroutine completed.")
				wg.Done()
			}(i, userCounterpartiesIDs)
		}
		if user.Data.InvitedByUserID != 0 {
			wg.Add(1)
			go func(i int, userID int64) {
				inviter, err := facade.User.GetUserByID(c, nil, userID)
				if err != nil {
					log.Errorf(c, fmt.Errorf("failed to get user by id=%v: %w", userID, err).Error())
					return
				}
				records[i].InvitedByUser = &struct {
					Id   int64
					Name string
				}{
					userID,
					inviter.Data.FullName(),
				}
				log.Debugf(c, "User goroutine completed.")
				wg.Done()
			}(i, user.Data.InvitedByUserID)
		}
	}

	wg.Wait()

	for i, record := range records {
		if userBytes, err := json.Marshal(record); err != nil {
			log.Errorf(c, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
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
		_, _ = w.Write([]byte(err.Error()))
	}
}
