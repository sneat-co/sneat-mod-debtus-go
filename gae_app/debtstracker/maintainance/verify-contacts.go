package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/strongo/db"
	"github.com/strongo/nds"
	"context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type verifyContacts struct {
	contactsAsyncJob
}

func (m *verifyContacts) Next(c context.Context, counters mapper.Counters, key *datastore.Key) error {
	//log.Debugf(c, "*verifyContacts.Next(id: %v)", key.IntID())
	return m.startContactWorker(c, counters, key, m.processContact)
}

func (m *verifyContacts) processContact(c context.Context, counters *asyncCounters, contact models.Contact) (err error) {
	if _, err = dal.User.GetUserByID(c, contact.UserID); db.IsNotFound(err) {
		counters.Increment("wrong_UserID", 1)
		log.Warningf(c, "Contact %d reference unknown user %d", contact.ID, contact.UserID)
	} else if err != nil {
		log.Errorf(c, err.Error())
		return
	}

	if err = m.verifyLinking(c, counters, contact); err != nil {
		return
	}

	if err = m.verifyBalance(c, counters, contact); err != nil {
		return
	}
	return
}

func (m *verifyContacts) verifyLinking(c context.Context, counters *asyncCounters, contact models.Contact) (err error) {
	if contact.CounterpartyCounterpartyID != 0 {
		var counterpartyContact models.Contact
		if counterpartyContact, err = dal.Contact.GetContactByID(c, contact.CounterpartyCounterpartyID); err != nil {
			log.Errorf(c, err.Error())
			return
		}
		if counterpartyContact.CounterpartyCounterpartyID == 0 || counterpartyContact.CounterpartyUserID == 0 {
			if err = m.linkContacts(c, counters, contact); err != nil {
				return
			}
		} else if counterpartyContact.CounterpartyCounterpartyID == contact.ID && counterpartyContact.CounterpartyUserID == contact.UserID {
			// Pass, we are OK
		} else {
			log.Warningf(c, "Wrongly linked contacts: %v=>%v != %v=>%v",
				contact.ID, contact.CounterpartyCounterpartyID,
				counterpartyContact.ID, counterpartyContact.CounterpartyCounterpartyID)
		}
	}
	return
}

func (m *verifyContacts) linkContacts(c context.Context, counters *asyncCounters, contact models.Contact) (err error) {
	var counterpartyContact models.Contact
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if counterpartyContact, err = dal.Contact.GetContactByID(c, contact.CounterpartyCounterpartyID); err != nil {
			log.Errorf(c, err.Error())
			return
		}
		if counterpartyContact.CounterpartyCounterpartyID == 0 {
			counterpartyContact.CounterpartyCounterpartyID = contact.ID
			if counterpartyContact.CounterpartyUserID == 0 {
				counterpartyContact.CounterpartyUserID = contact.UserID
			} else if counterpartyContact.CounterpartyUserID != contact.UserID {
				err = fmt.Errorf("counterpartyContact(id=%v).CounterpartyUserID != contact(id=%v).UserID: %v != %v",
					counterpartyContact.ID, contact.ID, counterpartyContact.CounterpartyUserID, contact.UserID)
				return
			}
			if err = dal.Contact.SaveContact(c, counterpartyContact); err != nil {
				return
			}
		} else if counterpartyContact.CounterpartyCounterpartyID != contact.ID {
			log.Warningf(c, "in tx: wrongly linked contacts: %v=>%v != %v=>%v",
				contact.ID, contact.CounterpartyCounterpartyID,
				counterpartyContact.ID, counterpartyContact.CounterpartyCounterpartyID)
		}
		return
	}, db.SingleGroupTransaction); err != nil {
		log.Errorf(c, err.Error())
		return
	}
	counters.Increment("linked_contacts", 1)
	log.Infof(c, "Successfully linked contact %v to %v", counterpartyContact.ID, contact.ID)
	return
}

func (m *verifyContacts) verifyBalance(c context.Context, counters *asyncCounters, contact models.Contact) (err error) {
	balance := contact.Balance()
	if FixBalanceCurrencies(balance) {
		if err = nds.RunInTransaction(c, func(c context.Context) (err error) {
			if contact, err = dal.Contact.GetContactByID(c, contact.ID); err != nil {
				return err
			}
			if balance := contact.Balance(); FixBalanceCurrencies(balance) {
				if err = contact.SetBalance(balance); err != nil {
					return err
				}
				if err = dal.Contact.SaveContact(c, contact); err != nil {
					return err
				}
				log.Infof(c, "Fixed contact balance currencies: %d", contact.ID)
			}
			return nil
		}, nil); err != nil {
			return
		}
	}
	return
}
