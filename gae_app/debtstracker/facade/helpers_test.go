package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal/dalmocks"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

func SetupMocks(c context.Context) dalmocks.MockDB {
	mockDB := dalmocks.NewMockDB()

	mockDB.UserMock.Users[1] = &models.AppUserEntity{ContactDetails: models.ContactDetails{FirstName: "Alfred", LastName: "Alpha"}}
	mockDB.UserMock.Users[3] = &models.AppUserEntity{ContactDetails: models.ContactDetails{FirstName: "Ben", LastName: "Bravo"}}
	mockDB.UserMock.Users[5] = &models.AppUserEntity{ContactDetails: models.ContactDetails{FirstName: "Charles", LastName: "Cain"}}

	dal.Contact.SaveContact(c, models.Contact{
		ID: 2,
		ContactEntity: &models.ContactEntity{
			Status:             models.STATUS_ACTIVE,
			UserID:             1,
			CounterpartyUserID: 3,
			ContactDetails:     models.ContactDetails{Nickname: "Bono"}},
	})
	dal.Contact.SaveContact(c, models.Contact{
		ID: 4,
		ContactEntity: &models.ContactEntity{
			Status:             models.STATUS_ACTIVE,
			UserID:             1,
			CounterpartyUserID: 5,
			ContactDetails:     models.ContactDetails{Nickname: "Carly"}},
	})
	dal.Contact.SaveContact(c, models.Contact{ID: 6, ContactEntity: &models.ContactEntity{
		Status: models.STATUS_ACTIVE, UserID: 1, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Den"}}})

	dal.Contact.SaveContact(c, models.Contact{ID: 8, ContactEntity: &models.ContactEntity{
		Status: models.STATUS_ACTIVE, UserID: 3, CounterpartyUserID: 1, ContactDetails: models.ContactDetails{Nickname: "Eagle"}}})
	dal.Contact.SaveContact(c, models.Contact{ID: 10, ContactEntity: &models.ContactEntity{
		Status: models.STATUS_ACTIVE, UserID: 5, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Ford"}}})
	dal.Contact.SaveContact(c, models.Contact{ID: 12, ContactEntity: &models.ContactEntity{
		Status: models.STATUS_ACTIVE, UserID: 5, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Gina"}}})

	dal.DB = mockDB
	return mockDB
}
