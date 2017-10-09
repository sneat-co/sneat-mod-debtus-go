package maintainance

import (
	"net/http"
	"github.com/captaincodeman/datastore-mapper"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"fmt"
	"github.com/strongo/app/log"
	"strings"
	"github.com/qedus/nds"
	"github.com/strongo/app/db"
)

type verifyTransfers struct {
	entity *models.TransferEntity
}

func (m *verifyTransfers) Query(r *http.Request) (*mapper.Query, error) {
	return mapper.NewQuery(models.TransferKind), nil
}

func (m *verifyTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	{
		for _, userID := range m.entity.BothUserIDs {
			if userID != 0 {
				if _, err2 := dal.User.GetUserByID(c, userID); db.IsNotFound(err2) {
					counters.Increment(fmt.Sprintf("User:%d", userID), 1)
					log.Warningf(c, "Transfer %d reference unknown user %d", key.IntID(), userID)
				} else if err2 != nil {
					err = err2
					return
				}
			}
		}
	}
	{
		for _, contactID := range m.entity.BothCounterpartyIDs {
			if contactID != 0 {
				if _, err2 := dal.Contact.GetContactByID(c, contactID); db.IsNotFound(err2) {
					counters.Increment(fmt.Sprintf("Contact:%d", contactID), 1)
					log.Warningf(c, "Transfer %d reference unknown contact %d", key.IntID(), contactID)
				} else if err2 != nil {
					err = err2
					return
				}
			}
		}
	}
	{
		var currency string
		if m.entity.Currency == "euro" {
			currency = "EUR"
		} else if len(m.entity.Currency) == 3 {
			v2 := strings.ToUpper(m.entity.Currency)
			if v2 != m.entity.Currency && models.Currency(v2).IsMoney() {
				currency = v2
			}
		}
		if currency != "" {
			if err = nds.RunInTransaction(c, func(c context.Context) error {
				if err = nds.Get(c, key, m.entity); err != nil {
					return err
				}
				if m.entity.Currency != currency {
					m.entity.Currency = currency
					if _, err = nds.Put(c, key, m.entity); err != nil {
						return err
					}
					log.Infof(c, "Transfer currency fixed: %d", key.IntID())
				}
				return nil
			}, nil); err != nil {
				return err
			}
		}
	}
	return
}

// JobStarted is called when a mapper job is started
func (m *verifyTransfers) JobStarted(c context.Context, id string) {

}

// JobStarted is called when a mapper job is completed
func (m *verifyTransfers) JobCompleted(c context.Context, id string) {
	logJobCompletion(c, id)
}

func (m *verifyTransfers) Make() interface{} {
	m.entity = new(models.TransferEntity)
	return m.entity
}
