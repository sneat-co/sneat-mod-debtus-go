package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"google.golang.org/appengine"
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

func (m *contactsAsyncJob) Query(r *http.Request) (query  *mapper.Query, err error) {
	c := appengine.NewContext(r)
	log.Debugf(c, "Query(): r.RawQuery: " + r.URL.RawQuery)
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