package gaedal

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/app/gae"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
	"google.golang.org/appengine/v2/delay"
)

type ContactDalGae struct {
}

func NewContactDalGae() ContactDalGae {
	return ContactDalGae{}
}

var _ dtdal.ContactDal = (*ContactDalGae)(nil)

func NewContactKey(c context.Context, contactID int64) *datastore.Key {
	if contactID == 0 {
		panic("NewContactKey(): contactID == 0")
	}
	return gaedb.NewKey(c, models.ContactKind, "", contactID, nil)
}

func NewContactIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.ContactKind, nil)
}

func (contactDalGae ContactDalGae) DeleteContact(c context.Context, contactID int64) (err error) {
	log.Debugf(c, "ContactDalGae.DeleteContact(%d)", contactID)
	if err = gaedb.Delete(c, NewContactKey(c, contactID)); err != nil {
		return
	}
	if err = delayDeleteContactTransfers(c, contactID, ""); err != nil { // TODO: Move to facade!
		return
	}
	return
}

var deleteContactTransfersDelayFunc *delay.Function

const DeleteContactTransfersFuncKey = "DeleteContactTransfers"

func init() {
	deleteContactTransfersDelayFunc = delay.Func(DeleteContactTransfersFuncKey, delayedDeleteContactTransfers)
}

func delayDeleteContactTransfers(c context.Context, contactID int64, cursor string) error {
	if err := gae.CallDelayFunc(c, common.QUEUE_TRANSFERS, DeleteContactTransfersFuncKey, deleteContactTransfersDelayFunc, contactID, cursor); err != nil {
		return err
	}
	return nil
}

func delayedDeleteContactTransfers(c context.Context, contactID int64, cursor string) (err error) {
	log.Debugf(c, "delayedDeleteContactTransfers(contactID=%d, cursor=%v", contactID, cursor)
	const limit = 100
	var transferIDs []int
	transferIDs, cursor, err = dtdal.Transfer.LoadTransferIDsByContactID(c, contactID, limit, cursor)
	if err != nil {
		return
	}
	keys := make([]*dal.Key, len(transferIDs))
	for i, transferID := range transferIDs {
		keys[i] = models.NewTransferKey(transferID)
	}
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
		if err = tx.DeleteMulti(c, keys); err != nil {
			return err
		}
		if len(transferIDs) == limit {
			if err = delayDeleteContactTransfers(c, contactID, cursor); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return
	}
	return
}

func (ContactDalGae) SaveContact(c context.Context, tx dal.ReadwriteTransaction, contact models.Contact) error {
	if err := tx.Set(c, contact.Record); err != nil {
		err = fmt.Errorf("failed to SaveContact(): %w", err)
	}
	return nil
}

func newContactQueryActive(userID int64) *datastore.Query {
	return newContactQueryWithStatus(userID, models.STATUS_ACTIVE)
}

func newContactQueryWithStatus(userID int64, status string) *datastore.Query {
	query := datastore.NewQuery(models.ContactKind).Filter("UserID =", userID)
	if status != "" {
		query = query.Filter("Status =", status)
	}
	return query
}

func (ContactDalGae) GetContactsWithDebts(c context.Context, userID int64) (counterparties []models.Contact, err error) {
	query := newContactQueryWithStatus(userID, "").Filter("BalanceCount >", 0)
	var (
		counterpartyKeys     []*datastore.Key
		counterpartyEntities []*models.ContactEntity
	)
	if counterpartyKeys, err = query.GetAll(c, &counterpartyEntities); err != nil {
		err = fmt.Errorf("method ContactDalGae.GetContactsWithDebts() failed to execute query.GetAll(): %w", err)
		return
	}
	counterparties = zipCounterparty(counterpartyKeys, counterpartyEntities)
	return
}

func (ContactDalGae) GetLatestContacts(whc botsfw.WebhookContext, limit, totalCount int) (counterparties []models.Contact, err error) {
	c := whc.Context()
	query := newContactQueryActive(whc.AppUserIntID()).Order("-LastTransferAt")
	if limit > 0 {
		query = query.Limit(limit)
	}
	var keys []*datastore.Key
	var entities []*models.ContactEntity
	if keys, err = query.GetAll(c, &entities); err != nil {
		err = fmt.Errorf("method ContactDalGae.GetLatestContacts() failed: %w", err)
		return
	}
	var contactsCount = len(keys)
	log.Debugf(c, "GetLatestContacts(limit=%v, totalCount=%v): %v", limit, totalCount, contactsCount)
	if (limit == 0 && contactsCount < totalCount) || (limit > 0 && totalCount > 0 && contactsCount < limit && contactsCount < totalCount) {
		log.Debugf(c, "Querying counterparties without index -LastTransferAt")
		query = newContactQueryActive(whc.AppUserIntID())
		if limit > 0 {
			query = query.Limit(limit)
		}
		if keys2, err := query.GetAll(c, &entities); err != nil {
			err = fmt.Errorf("ContactDalGae.GetLatestContacts() failed: %w", err)
			return nil, err
		} else {
			keys = append(keys, keys2...)
		}
		log.Debugf(c, "len(keys): %v, len(entities): %v", len(keys), len(entities))
	}
	counterparties = zipCounterparty(keys, entities)
	return
}

func (contactDalGae ContactDalGae) GetContactIDsByTitle(c context.Context, tx dal.ReadTransaction, userID int64, title string, caseSensitive bool) (contactIDs []int64, err error) {
	var user models.AppUser
	if user, err = facade.User.GetUserByID(c, userID); err != nil {
		return
	}
	if caseSensitive {
		for _, contact := range user.Data.Contacts() {
			if contact.Name == title {
				contactIDs = append(contactIDs, contact.ID)
			}
		}
	} else {
		title = strings.ToLower(title)
		for _, contact := range user.Data.Contacts() {
			if strings.ToLower(contact.Name) == title {
				contactIDs = append(contactIDs, contact.ID)
			}
		}
	}
	return
}

func zipCounterparty(keys []*datastore.Key, entities []*models.ContactEntity) (contacts []models.Contact) {
	if len(keys) != len(entities) {
		panic(fmt.Sprintf("len(keys):%d != len(entities):%d", len(keys), len(entities)))
	}
	contacts = make([]models.Contact, len(entities))
	for i, entity := range entities {
		contacts[i] = models.NewContact(keys[i].IntID(), entity)
	}
	return
}

func (contactDalGae ContactDalGae) InsertContact(c context.Context, tx dal.ReadwriteTransaction, contactEntity *models.ContactEntity) (
	contact models.Contact, err error,
) {
	contact.Data = contactEntity
	if err = tx.Insert(c, contact.Record); err != nil {
		return
	}
	contact.ID = contact.Key.ID.(int64)
	return
}
