package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
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
