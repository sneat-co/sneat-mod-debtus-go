package dalmocks

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/db"
	"golang.org/x/net/context"
)

var _ dal.ContactDal = (*ContactDalMock)(nil)

type ContactDalMock struct {
	LastContactID int64
	Contacts      map[int64]*models.ContactEntity
}

func NewContactDalMock() *ContactDalMock {
	return &ContactDalMock{Contacts: make(map[int64]*models.ContactEntity)}
}

func (mock *ContactDalMock) GetContactByID(c context.Context, contactID int64) (contact models.Contact, err error) {
	if contactEntity, ok := mock.Contacts[contactID]; ok {
		contact.ContactEntity = contactEntity
		contact.ID = contactID
	} else {
		err = db.ErrRecordNotFound
	}
	return
}

func (mock *ContactDalMock) GetContactsByIDs(c context.Context, contactsIDs []int64) (contacts []models.Contact, err error) {
	contacts = make([]models.Contact, len(contactsIDs))
	var found bool
	for i, id := range contactsIDs {
		contact := contacts[i]
		contact.ID = id
		if contact.ContactEntity, found = mock.Contacts[id]; !found {
			err = db.NewErrNotFoundByIntID(models.ContactKind, id, nil)
			return
		}
		contacts[i] = contact
	}
	return
}

func (mock *ContactDalMock) GetLatestContacts(whc bots.WebhookContext, limit, totalCount int) (contacts []models.Contact, err error) {
	return
}

func (mock *ContactDalMock) InsertContact(c context.Context, contactEntity *models.ContactEntity) (contact models.Contact, err error) {
	if contactEntity == nil {
		panic("contactEntity == nil")
	}
	mock.LastContactID += 1
	contact.ID = mock.LastContactID
	contact.ContactEntity = contactEntity
	mock.Contacts[mock.LastContactID] = contact.ContactEntity
	return
}

//CreateContact(c context.Context, userID int64, contactDetails models.ContactDetails) (contact models.Contact, user models.AppUser, err error)
//CreateContactWithinTransaction(c context.Context, user models.AppUser, contactUserID, counterpartyCounterpartyID int64, contactDetails models.ContactDetails, balanced models.Balanced) (contact models.Contact, err error)
//UpdateContact(c context.Context, contactID int64, values map[string]string) (contactEntity *models.ContactEntity, err error)

func (mock *ContactDalMock) SaveContact(c context.Context, contact models.Contact) (err error) {
	mock.Contacts[contact.ID] = contact.ContactEntity
	return
}

func (mock *ContactDalMock) DeleteContact(c context.Context, contactID int64) (err error) {
	delete(mock.Contacts, contactID)
	return
}

func (mock *ContactDalMock) GetContactIDsByTitle(c context.Context, userID int64, title string, caseSensitive bool) (contactIDs []int64, err error) {
	return
}

func (mock *ContactDalMock) GetContactsWithDebts(c context.Context, userID int64) (contacts []models.Contact, err error) {
	return
}
