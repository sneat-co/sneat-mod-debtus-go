package dalmocks

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/bots-go-framework/bots-fw/botsfw"
)

var _ dtdal.ContactDal = (*ContactDalMock)(nil)

type ContactDalMock struct {
	LastContactID int64
	Contacts      map[int64]*models.ContactEntity
}

func NewContactDalMock() *ContactDalMock {
	return &ContactDalMock{Contacts: make(map[int64]*models.ContactEntity)}
}

func (mock *ContactDalMock) GetLatestContacts(whc botsfw.WebhookContext, limit, totalCount int) (contacts []models.Contact, err error) {
	return
}

func (mock *ContactDalMock) InsertContact(c context.Context, contactEntity *models.ContactEntity) (contact models.Contact, err error) {
	if contactEntity == nil {
		panic("contactEntity == nil")
	}
	mock.LastContactID += 1
	contact.ID = mock.LastContactID
	contact.Data = contactEntity
	mock.Contacts[mock.LastContactID] = contact.Data
	return
}

//CreateContact(c context.Context, userID int64, contactDetails models.ContactDetails) (contact models.Contact, user models.AppUser, err error)
//CreateContactWithinTransaction(c context.Context, user models.AppUser, contactUserID, counterpartyCounterpartyID int64, contactDetails models.ContactDetails, balanced money.Balanced) (contact models.Contact, err error)
//UpdateContact(c context.Context, contactID int64, values map[string]string) (contactEntity *models.ContactEntity, err error)

func (mock *ContactDalMock) SaveContact(c context.Context, contact models.Contact) (err error) {
	mock.Contacts[contact.ID] = contact.Data
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
