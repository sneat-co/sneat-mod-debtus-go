package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"encoding/json"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bytes"
	"fmt"
	"github.com/strongo/app/db"
)

type verifyUsers struct {
	asyncMapper
	entity *models.AppUserEntity
}

func (m *verifyUsers) Make() interface{} {
	m.entity = new(models.AppUserEntity)
	return m.entity
}

func (m *verifyUsers) Query(r *http.Request) (query *mapper.Query, err error) {
	return filterByIntID(r, models.AppUserKind,"user")
}

func (m *verifyUsers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	userEntity := *m.entity
	user := models.AppUser{IntegerID: db.NewIntID(key.IntID()), AppUserEntity: &userEntity}
	return m.startWorker(c, counters, func() Worker {
		return func(counters *asyncCounters) error {
			return m.processUser(c, user, counters)
		}
	})
}

func (m *verifyUsers) processUser(c context.Context, user models.AppUser, counters *asyncCounters) (err error) {
	buf := new(bytes.Buffer)
	if err = m.verifyUserBalanceAndContacts(c, buf, counters, user); err != nil {
		return
	}
	if buf.Len() > 0 {
		log.Infof(c, buf.String())
	}
	return
}

func (m *verifyUsers) verifyUserBalanceAndContacts(c context.Context, buf *bytes.Buffer, counters *asyncCounters, user models.AppUser) (err error) {
	if user.BalanceCount > 0 {
		balance := user.Balance()

		if fixedContactsBalances, err := fixUserContactsBalances(m.entity); err != nil {
			return err
		} else if fixedContactsBalances || FixBalanceCurrencies(balance) {
			if err = nds.RunInTransaction(c, func(c context.Context) error {
				if user, err = dal.User.GetUserByID(c, user.ID); err != nil {
					return err
				}
				balance = m.entity.Balance()
				if err != nil {
					return err
				}
				changed := false
				if FixBalanceCurrencies(balance) {
					m.entity.SetBalance(balance)
					changed = true
				}
				if fixedContactsBalances, err = fixUserContactsBalances(m.entity); err != nil {
					return err
				} else if fixedContactsBalances {
					changed = true
				}
				if changed {
					if err = dal.User.SaveUser(c, user); err != nil {
						return err
					}
					fmt.Fprintf(buf, "User fixed: %d ", user.ID)
				}
				return nil
			}, nil); err != nil {
				return err
			}
		}
	}
	return
}

func fixUserContactsBalances(u *models.AppUserEntity) (changed bool, err error) {
	contacts := u.Contacts()
	for i, contact := range contacts {
		if balance := contact.Balance(); FixBalanceCurrencies(balance) {
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
