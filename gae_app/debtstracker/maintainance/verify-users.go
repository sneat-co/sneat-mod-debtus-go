package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"encoding/json"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
	"strings"
)

type verifyUsers struct {
	entity *models.AppUserEntity
}

func (m *verifyUsers) Query(r *http.Request) (*mapper.Query, error) {
	return mapper.NewQuery(models.AppUserKind), nil
}

func (m *verifyUsers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	if err = m.verifyUserBalanceAndContacts(c, counters, key); err != nil {
		return
	}
	return nil
}

func (m *verifyUsers) verifyUserBalanceAndContacts(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	if m.entity.BalanceCount > 0 {
		balance, err := m.entity.Balance()
		if err != nil {
			return err
		}

		if fixedContactsBalances, err := FixUserContactsBalances(m.entity); err != nil {
			return err
		} else if fixedContactsBalances || FixBalanceCurrencies(balance) {
			if err = nds.RunInTransaction(c, func(c context.Context) error {
				if err = nds.Get(c, key, m.entity); err != nil {
					return err
				}
				balance, err = m.entity.Balance()
				if err != nil {
					return err
				}
				changed := false
				if FixBalanceCurrencies(balance) {
					m.entity.SetBalance(balance)
					changed = true
				}
				if fixedContactsBalances, err = FixUserContactsBalances(m.entity); err != nil {
					return err
				} else if fixedContactsBalances {
					changed = true
				}
				if changed {
					if _, err = nds.Put(c, key, m.entity); err != nil {
						return err
					}
					log.Infof(c, "User fixed: %d ", key.IntID())
				}
				return nil
			}, nil); err != nil {
				return err
			}
		}
	}
	return
}

func FixUserContactsBalances(u *models.AppUserEntity) (changed bool, err error) {
	contacts := u.Contacts()
	for i, contact := range contacts {
		var balance models.Balance
		if balance, err = contact.Balance(); err != nil {
			return
		}
		if FixBalanceCurrencies(balance) {
			balanceJsonBytes, err := ffjson.Marshal(balance)
			if err != nil {
				return changed, err
			}
			balanceJson := json.RawMessage(balanceJsonBytes)
			contact.BalanceJson = &balanceJson
			contacts[i] = contact
			changed = true
		}
	}
	if changed {
		u.SetContacts(contacts)
	}
	return
}

func FixBalanceCurrencies(balance models.Balance) (changed bool) {
	euro := models.Currency("euro")
	for c, v := range balance {
		if c == euro {
			c = models.CURRENCY_EUR
		} else if len(c) == 3 {
			cc := strings.ToUpper(string(c))
			if cc != string(c) {
				if cu := models.Currency(cc); cu.IsMoney() {
					balance[cu] += v
					delete(balance, c)
					changed = true
				}
			}
		}
	}
	return
}

func (m *verifyUsers) Make() interface{} {
	m.entity = new(models.AppUserEntity)
	return m.entity
}

// JobStarted is called when a mapper job is started
func (m *verifyUsers) JobStarted(c context.Context, id string) {
	log.Debugf(c, "Job started: %v", id)
}

// JobStarted is called when a mapper job is completed
func (m *verifyUsers) JobCompleted(c context.Context, id string) {
	logJobCompletion(c, id)
}
