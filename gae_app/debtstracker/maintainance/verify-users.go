package maintainance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/pkg/errors"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/db"
	"github.com/strongo/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"time"
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
	var filtered bool
	if query, filtered, err = filterByIntID(r, models.AppUserKind, "user"); err != nil {
		return
	} else if filtered {
		if len(r.URL.Query()) != 1 {
			err = errors.New("unexpected params: " + r.URL.RawQuery)
		}
		return
	}
	return
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
	if user, err = m.checkContactsExistsAndRecreateIfNeeded(c, buf, counters, user); err != nil {
		return
	}
	if err = m.verifyUserBalanceAndContacts(c, buf, counters, user); err != nil {
		return
	}
	if buf.Len() > 0 {
		log.Infof(c, buf.String())
	}
	return
}

func (m *verifyUsers) checkContactsExistsAndRecreateIfNeeded(c context.Context, buf *bytes.Buffer, counters *asyncCounters, user models.AppUser) (models.AppUser, error) {
	userContacts := user.Contacts()
	userChanged := false
	var err error
	for i, userContact := range userContacts {
		contactID := userContact.ID
		var contact models.Contact
		if contact, err = dal.Contact.GetContactByID(c, contactID); err != nil {
			if db.IsNotFound(err) {
				if err = m.createContact(c, buf, counters, user, userContact); err != nil {
					log.Errorf(c, "Failed to create contact %v", userContact.ID)
					err = nil
					continue
				}
			} else {
				return user, err
			}
		}
		if contact.CounterpartyUserID != 0 && userContact.UserID != contact.CounterpartyUserID {
			if userContact.UserID == 0 {
				userContact.UserID = contact.CounterpartyUserID
				userContacts[i] = userContact
				userChanged = true
			} else {
				err = fmt.Errorf(
					"data integrity issue for contact %v: userContact.UserID != contact.CounterpartyUserID: %v != %v",
					contact.ID, userContact.UserID, contact.CounterpartyUserID)
				return user, err
			}
		}
	}
	if userChanged {
		if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
			if user, err = dal.User.GetUserByID(c, user.ID); err != nil {
				return err
			}
			user.SetContacts(userContacts)
			if err = dal.User.SaveUser(c, user); err != nil {
				return err
			}
			return nil
		}, db.CrossGroupTransaction); err != nil {
			return user, err
		}

	}
	return user, err
}

func (m *verifyUsers) createContact(c context.Context, buf *bytes.Buffer, counters *asyncCounters, user models.AppUser, userContact models.UserContactJson) (err error) {
	var contact models.Contact
	if err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
		if contact, err = dal.Contact.GetContactByID(tc, userContact.ID); err != nil {
			if db.IsNotFound(err) {
				contact = models.NewContact(userContact.ID, &models.ContactEntity{
					UserID: user.ID,
					DtCreated: time.Now(),
					Status: models.STATUS_ACTIVE,
					ContactDetails: models.ContactDetails{
						Nickname:       userContact.Name,
						TelegramUserID: userContact.TgUserID,
					},
				})
				if err = contact.SetBalance(userContact.Balance()); err != nil {
					return
				}
				if err = contact.SetTransfersInfo(*contact.GetTransfersInfo()); err != nil {
					return
				}
				if err = dal.Contact.SaveContact(tc, contact); err != nil {
					return
				}
			}
			return
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		return
	} else {
		log.Warningf(c, "Recreated contact %v[%v] for user %v[%v]", contact.ID, contact.FullName(), user.ID, user.FullName())
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
