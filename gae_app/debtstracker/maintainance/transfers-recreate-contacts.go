package maintainance

import (
	"fmt"
	"runtime/debug"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"google.golang.org/appengine/datastore"
	"time"
)

type transfersRecreateContacts struct {
	transfersAsyncJob
}

func (m *transfersRecreateContacts) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	return m.startTransferWorker(c, counters, key, m.verifyAndFix)
}

func (m *transfersRecreateContacts) verifyAndFix(c context.Context, counters *asyncCounters, transfer models.Transfer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf(c, "*transfersRecreateContacts.verifyAndFix() => panic: %v\n\n%v", r, string(debug.Stack()))
		}
	}()
	var fixed bool
	fixed, err = verifyAndFixMissingTransferContacts(c, transfer)
	if fixed {
		counters.Increment("fixed", 1)
	}
	return
}

func verifyAndFixMissingTransferContacts(c context.Context, transfer models.Transfer) (fixed bool, err error) {
	isMissingAndCanBeFixed := func(contactID, contactUserID, counterpartyContactID int64) (bool, error) {
		if contactID != 0 && contactUserID != 0 && counterpartyContactID != 0 {
			if _, err := dal.Contact.GetContactByID(c, contactID); err != nil {
				if db.IsNotFound(err) {
					if user, err := dal.User.GetUserByID(c, contactUserID); err != nil {
						return false, err
					} else {
						for _, c := range user.Contacts() {
							if c.ID == contactID {
								return true, nil
							}
						}
						return false, nil
					}
				}
				return false, err
			}
		}
		return false, nil
	}

	doFix := func(contactInfo *models.TransferCounterpartyInfo, counterpartyInfo *models.TransferCounterpartyInfo) (err error) {
		err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
			log.Debugf(c, "Recreating contact # %v", contactInfo.ContactID)
			var counterpartyContact models.Contact
			if counterpartyContact, err = dal.Contact.GetContactByID(c, counterpartyInfo.ContactID); err != nil {
				return
			}
			var contactUser, counterpartyUser models.AppUser

			if contactUser, err = dal.User.GetUserByID(c, counterpartyInfo.UserID); err != nil {
				return
			}

			if counterpartyUser, err = dal.User.GetUserByID(c, contactInfo.UserID); err != nil {
				return
			}

			var contactUserContactJson models.UserContactJson

			for _, c := range contactUser.Contacts() {
				if c.ID == contactInfo.ContactID {
					contactUserContactJson = c
					break
				}
			}

			if contactUserContactJson.ID == 0 {
				log.Errorf(c, "Contact %v info not found in user %v contacts json", contactInfo.ContactID, counterpartyInfo.UserID)
				return
			}

			if counterpartyContact.CounterpartyCounterpartyID == 0 {
				if counterpartyContact.CounterpartyCounterpartyID == 0 {
					counterpartyContact.CounterpartyCounterpartyID = contactInfo.ContactID
					counterpartyContact.CounterpartyUserID = counterpartyInfo.UserID
				} else if counterpartyContact.CounterpartyCounterpartyID != contactInfo.ContactID {
					log.Errorf(c, "counterpartyContact.CounterpartyCounterpartyID != contact.ID: %v != %v", counterpartyContact.CounterpartyCounterpartyID, contactInfo.ContactID)
					return
				}
				if err = dal.Contact.SaveContact(c, counterpartyContact); err != nil {
					return err
				}
			}

			contact := models.NewContact(contactInfo.ContactID, &models.ContactEntity{
				UserID:         counterpartyInfo.UserID,
				DtCreated:      time.Now(),
				Status:         models.STATUS_ACTIVE,
				TransfersJson:  counterpartyContact.TransfersJson,
				ContactDetails: counterpartyUser.ContactDetails,
				Balanced:       counterpartyContact.Balanced,
			})
			if contact.Nickname != contactUserContactJson.Name && contact.FirstName != contactUserContactJson.Name && contact.LastName != contactUserContactJson.Name && contact.ScreenName != contactUserContactJson.Name {
				contact.Nickname = contactUserContactJson.Name
			}
			if err = contact.SetBalance(counterpartyContact.Balance().Reversed()); err != nil {
				return
			}
			if !contact.Balance().Equal(contactUserContactJson.Balance()) {
				err = fmt.Errorf("contact(%v).Balance != contactUserContactJson.Balance(): %v != %v", contact.ID, contact.Balance(), contactUserContactJson.Balance())
				return
			}
			if err = dal.Contact.SaveContact(c, contact); err != nil {
				return
			}

			return
		}, db.CrossGroupTransaction)
		if err != nil {
			return
		}
		fixed = true
		log.Warningf(c, "Counterparty re-created: %v", contactInfo.ContactID)
		return
	}

	verifyAndFix := func(contactInfo *models.TransferCounterpartyInfo, counterpartyInfo *models.TransferCounterpartyInfo) error {
		if toBeFixed, err := isMissingAndCanBeFixed(contactInfo.ContactID, counterpartyInfo.UserID, counterpartyInfo.ContactID); err != nil {
			return err
		} else if toBeFixed {
			return doFix(contactInfo, counterpartyInfo)
		}
		return nil
	}

	from, to := transfer.From(), transfer.To()

	if err = verifyAndFix(from, to); err != nil {
		return
	}

	if err = verifyAndFix(to, from); err != nil {
		return
	}
	return
}
