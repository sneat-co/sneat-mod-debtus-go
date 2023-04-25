package maintainance

import (
	"net/http"

	"context"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"google.golang.org/appengine/datastore"
)

type contactsAsyncJob struct {
	asyncMapper
	entity *models.ContactData
}

var _ mapper.JobEntity = (*contactsAsyncJob)(nil)

func (m *contactsAsyncJob) Make() interface{} {
	m.entity = new(models.ContactData)
	return m.entity
}

func (m *contactsAsyncJob) Query(r *http.Request) (query *mapper.Query, err error) {
	return applyIDAndUserFilters(r, "contactsAsyncJob", models.ContactKind, filterByIntID, "UserID")
}

func (m *contactsAsyncJob) Contact(key *datastore.Key) (contact models.Contact) {
	contact = models.NewContact(key.IntID(), nil)
	if m.entity != nil {
		entity := *m.entity
		contact.Data = &entity
	}
	return
}

type ContactWorker func(c context.Context, counters *asyncCounters, contact models.Contact) error

//func (m *contactsAsyncJob) startContactWorker(c context.Context, counters mapper.Counters, key *datastore.Key, contactWorker ContactWorker) error {
//	//log.Debugf(c, "*contactsAsyncJob.startContactWorker()")
//	contact := m.Contact(key)
//	createContactWorker := func() Worker {
//		//log.Debugf(c, "createContactWorker()")
//		return func(counters *asyncCounters) error {
//			//log.Debugf(c, "asyncContactWorker() => contact.ID: %v", contact.ID)
//			return contactWorker(c, counters, contact)
//		}
//	}
//	return m.startWorker(c, counters, createContactWorker)
//}
