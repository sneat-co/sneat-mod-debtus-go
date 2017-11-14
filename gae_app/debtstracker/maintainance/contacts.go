package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
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
	if query, err = filterByIntID(r, models.ContactKind, "contact"); err != nil {
		return
	}
	return filterByUserParam(r, mapper.NewQuery(models.ContactKind), "UserID")
}

func (m *contactsAsyncJob) Contact(key *datastore.Key) models.Contact {
	entity := *m.entity
	return models.NewContact(key.IntID(), &entity)
}

type ContactWorker func(c context.Context, counters *asyncCounters, contact models.Contact) error

func (m *contactsAsyncJob) startContactWorker(c context.Context, counters mapper.Counters, key *datastore.Key, contactWorker ContactWorker) error {
	contact := m.Contact(key)
	worker := func() Worker {
		return func(counters *asyncCounters) error {
			return contactWorker(c, counters, contact)
		}
	}
	return m.startWorker(c, counters, worker)
}