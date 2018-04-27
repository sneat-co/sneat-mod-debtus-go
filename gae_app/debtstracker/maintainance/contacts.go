package maintainance

import (
	"net/http"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type contactsAsyncJob struct {
	asyncMapper
	entity *models.ContactEntity
}

var _ mapper.JobEntity = (*contactsAsyncJob)(nil)

func (m *contactsAsyncJob) Make() interface{} {
	m.entity = new(models.ContactEntity)
	return m.entity
}

func (m *contactsAsyncJob) Query(r *http.Request) (query *mapper.Query, err error) {
	c := appengine.NewContext(r)
	log.Debugf(c, "Query(): r.RawQuery: "+r.URL.RawQuery)
	var filtered bool
	if query, filtered, err = filterByIntID(r, models.ContactKind, "contact"); err != nil {
		log.Errorf(c, err.Error())
		return
	} else if filtered {
		return
	}
	paramsCount := len(r.URL.Query()) - 1 // 1 parameter is job name
	if query, filtered, err = filterByUserParam(r, mapper.NewQuery(models.ContactKind), "UserID"); err != nil {
		log.Errorf(c, err.Error())
		return
	} else if filtered {
		paramsCount -= 1
	}

	if paramsCount > 0 {
		err = errors.New("Some unknown parameters: " + r.URL.RawQuery)
		log.Errorf(c, err.Error())
		return
	}
	return
}

func (m *contactsAsyncJob) Contact(key *datastore.Key) (contact models.Contact) {
	contact = models.NewContact(key.IntID(), nil)
	if m.entity != nil {
		entity := *m.entity
		contact.ContactEntity = &entity
	}
	return
}

type ContactWorker func(c context.Context, counters *asyncCounters, contact models.Contact) error

func (m *contactsAsyncJob) startContactWorker(c context.Context, counters mapper.Counters, key *datastore.Key, contactWorker ContactWorker) error {
	//log.Debugf(c, "*contactsAsyncJob.startContactWorker()")
	contact := m.Contact(key)
	createContactWorker := func() Worker {
		//log.Debugf(c, "createContactWorker()")
		return func(counters *asyncCounters) error {
			//log.Debugf(c, "asyncContactWorker() => contact.ID: %v", contact.ID)
			return contactWorker(c, counters, contact)
		}
	}
	return m.startWorker(c, counters, createContactWorker)
}
