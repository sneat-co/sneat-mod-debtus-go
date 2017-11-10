package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
)

type contactsAsyncJob struct {
	asyncMapper
	entity *models.ContactEntity
}

func (m *contactsAsyncJob) Make() interface{} {
	m.entity = new(models.ContactEntity)
	return m.entity
}


func (m *contactsAsyncJob) Query(r *http.Request) (query  *mapper.Query, err error) {
	return filterByUserParam(r, mapper.NewQuery(models.ContactKind), "UserID")
}


