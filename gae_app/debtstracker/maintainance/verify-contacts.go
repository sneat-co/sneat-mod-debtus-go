package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/strongo/nds"
	"github.com/strongo/db"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type verifyContacts struct {
	contactsAsyncJob
}

func (m *verifyContacts) Next(c context.Context, counters mapper.Counters, key *datastore.Key) error {
	return m.startContactWorker(c, counters, key, m.processContact)
}

func (m *verifyContacts) processContact(c context.Context, counters *asyncCounters, contact models.Contact) (err error) {
	//buf := new(bytes.Buffer)

	if _, err = dal.User.GetUserByID(c, m.entity.UserID); db.IsNotFound(err) {
		counters.Increment(fmt.Sprintf("User:%d", m.entity.UserID), 1)
		log.Warningf(c, "Contact %d reference unknown user %d", contact.ID, m.entity.UserID)
	} else if err != nil {
		log.Errorf(c, err.Error())
		return
	}
	balance := m.entity.Balance()
	if FixBalanceCurrencies(balance) {
		if err = nds.RunInTransaction(c, func(c context.Context) (err error) {
			if contact, err = dal.Contact.GetContactByID(c, contact.ID); err != nil {
				return err
			}
			if balance := m.entity.Balance(); FixBalanceCurrencies(balance) {
				if err = m.entity.SetBalance(balance); err != nil {
					return err
				}
				if err = dal.Contact.SaveContact(c, contact); err != nil {
					return err
				}
				log.Infof(c, "Contact fixed: %d", contact.ID)
			}
			return nil
		}, nil); err != nil {
			return
		}
	}
	return
}

