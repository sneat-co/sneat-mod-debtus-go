package gaedal

import (
	"context"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/common"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/models"
	"github.com/strongo/delaying"
	"github.com/strongo/log"
	"strings"
)

type ContactDalGae struct {
}

func NewContactDalGae() ContactDalGae {
	return ContactDalGae{}
}

var _ dtdal.ContactDal = (*ContactDalGae)(nil)

func (contactDalGae ContactDalGae) DeleteContact(c context.Context, tx dal.ReadwriteTransaction, contactID string) (err error) {
	log.Debugf(c, "ContactDalGae.DeleteContact(%d)", contactID)
	if err = tx.Delete(c, models.NewContactKey(contactID)); err != nil {
		return
	}
	if err = delayDeleteContactTransfers(c, contactID, ""); err != nil { // TODO: Move to facade!
		return
	}
	return
}

const DeleteContactTransfersFuncKey = "DeleteContactTransfers"

func delayDeleteContactTransfers(c context.Context, contactID string, cursor string) error {
	if err := delayDeleteContactTransfersDelayFunc.EnqueueWork(c, delaying.With(common.QUEUE_TRANSFERS, DeleteContactTransfersFuncKey, 0), contactID, cursor); err != nil {
		return err
	}
	return nil
}

func delayedDeleteContactTransfers(c context.Context, contactID string, cursor string) (err error) {
	log.Debugf(c, "delayedDeleteContactTransfers(contactID=%d, cursor=%v", contactID, cursor)
	const limit = 100
	var transferIDs []string
	transferIDs, cursor, err = dtdal.Transfer.LoadTransferIDsByContactID(c, contactID, limit, cursor)
	if err != nil {
		return
	}
	keys := make([]*dal.Key, len(transferIDs))
	for i, transferID := range transferIDs {
		keys[i] = models.NewTransferKey(transferID)
	}
	var db dal.DB
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
		return fmt.Errorf("failed to SaveContact(): %w", err)
	}
	return nil
}

func newUserActiveContactsQuery(userID string) dal.QueryBuilder {
	return newUserContactsQuery(userID).WhereField("Status", dal.Equal, models.STATUS_ACTIVE)
}

func newUserContactsQuery(userID string) dal.QueryBuilder {
	return dal.From(models.ContactKind).WhereField("UserID", dal.Equal, userID)
}

func (ContactDalGae) GetContactsWithDebts(c context.Context, tx dal.ReadSession, userID string) (counterparties []models.Contact, err error) {
	query := newUserContactsQuery(userID).
		WhereField("BalanceCount", dal.GreaterThen, 0).
		SelectInto(models.NewContactRecord)
	//var (
	//	counterpartyEntities []*models.ContactData
	//)
	records, err := tx.QueryAllRecords(c, query)
	counterparties = make([]models.Contact, len(records))
	for i, record := range records {
		counterparties[i] = models.NewContact(record.Key().ID.(string), record.Data().(*models.ContactData))
	}
	return
}

func (ContactDalGae) GetLatestContacts(whc botsfw.WebhookContext, tx dal.ReadSession, limit, totalCount int) (counterparties []models.Contact, err error) {
	c := whc.Context()
	appUserID := whc.AppUserID()
	query := newUserActiveContactsQuery(appUserID).
		OrderBy(dal.DescendingField("LastTransferAt")).
		Limit(limit).
		SelectInto(models.NewContactRecord)
	if tx == nil {
		if tx, err = facade.GetDatabase(c); err != nil {
			return
		}
	}
	var records []dal.Record
	records, err = tx.QueryAllRecords(c, query)
	var contactsCount = len(records)
	log.Debugf(c, "GetLatestContacts(limit=%v, totalCount=%v): %v", limit, totalCount, contactsCount)
	if (limit == 0 && contactsCount < totalCount) || (limit > 0 && totalCount > 0 && contactsCount < limit && contactsCount < totalCount) {
		log.Debugf(c, "Querying counterparties without index -LastTransferAt")
		query = newUserActiveContactsQuery(appUserID).
			Limit(limit).
			SelectInto(models.NewTransferRecord)
		if records, err = tx.QueryAllRecords(c, query); err != nil {
			return
		}
	}
	counterparties = make([]models.Contact, len(records))
	for i, record := range records {
		counterparties[i] = models.NewContact(record.Key().ID.(string), record.Data().(*models.ContactData))
	}
	return
}

func (contactDalGae ContactDalGae) GetContactIDsByTitle(c context.Context, tx dal.ReadSession, userID string, title string, caseSensitive bool) (contactIDs []string, err error) {
	var user models.AppUser
	if user, err = facade.User.GetUserByID(c, tx, userID); err != nil {
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

//func zipCounterparty(keys []*datastore.Key, entities []*models.ContactData) (contacts []models.Contact) {
//	if len(keys) != len(entities) {
//		panic(fmt.Sprintf("len(keys):%d != len(entities):%d", len(keys), len(entities)))
//	}
//	contacts = make([]models.Contact, len(entities))
//	for i, entity := range entities {
//		contacts[i] = models.NewContact(keys[i].IntID(), entity)
//	}
//	return
//}

func (contactDalGae ContactDalGae) InsertContact(c context.Context, tx dal.ReadwriteTransaction, contactEntity *models.ContactData) (
	contact models.Contact, err error,
) {
	contact.Data = contactEntity
	if err = tx.Insert(c, contact.Record); err != nil {
		return
	}
	contact.ID = contact.Key.ID.(string)
	return
}
