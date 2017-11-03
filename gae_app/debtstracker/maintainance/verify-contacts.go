package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/strongo/nds"
	"github.com/strongo/app/db"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
)

type verifyContacts struct {
	entity *models.ContactEntity
}

func (m *verifyContacts) Query(r *http.Request) (*mapper.Query, error) {
	return mapper.NewQuery(models.ContactKind), nil
}

func (m *verifyContacts) Next(c context.Context, counters mapper.Counters, key *datastore.Key) error {
	if _, err := dal.User.GetUserByID(c, m.entity.UserID); db.IsNotFound(err) {
		counters.Increment(fmt.Sprintf("User:%d", m.entity.UserID), 1)
		log.Warningf(c, "Contact %d reference unknown user %d", key.IntID(), m.entity.UserID)
	} else if err != nil {
		return err
	}
	balance, err := m.entity.Balance()
	if err != nil {
		return err
	}
	if FixBalanceCurrencies(balance) {
		if err = nds.RunInTransaction(c, func(c context.Context) error {
			if err = nds.Get(c, key, m.entity); err != nil {
				return err
			}
			balance, err := m.entity.Balance()
			if err != nil {
				return err
			}
			if FixBalanceCurrencies(balance) {
				m.entity.SetBalance(balance)
				if _, err = nds.Put(c, key, m.entity); err != nil {
					return err
				}
				log.Infof(c, "Contact fixed: %d", key.IntID())
			}
			return nil
		}, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *verifyContacts) Make() interface{} {
	m.entity = new(models.ContactEntity)
	return m.entity
}

// JobStarted is called when a mapper job is started
func (m *verifyContacts) JobStarted(c context.Context, id string) {
	log.Debugf(c, "Job started: %v", id)
}

// JobStarted is called when a mapper job is completed
func (m *verifyContacts) JobCompleted(c context.Context, id string) {
	logJobCompletion(c, id)
}
