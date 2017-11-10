package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type transfersAsyncJob struct {
	asyncMapper
	entity *models.TransferEntity
}

func (m *transfersAsyncJob) Make() interface{} {
	m.entity = new(models.TransferEntity)
	return m.entity
}


func (m *transfersAsyncJob) Query(r *http.Request) (query  *mapper.Query, err error) {
	return filterByUserParam(r, mapper.NewQuery(models.TransferKind), "BothUserIDs")
}

func (m *transfersAsyncJob) Transfer(key *datastore.Key) models.Transfer {
	entity := *m.entity
	return models.Transfer{ID: key.IntID(), TransferEntity: &entity}
}


type TransferWorker func(c context.Context, counters *asyncCounters, transfer models.Transfer) error

func (m *transfersAsyncJob) startTransferWorker(c context.Context, counters mapper.Counters, key *datastore.Key, transferWorker TransferWorker) error {
	transfer := m.Transfer(key)
	w := func() Worker {
		return func(counters *asyncCounters) error {
			return transferWorker(c, counters, transfer)
		}
	}
	return m.startWorker(c, counters, w)
}