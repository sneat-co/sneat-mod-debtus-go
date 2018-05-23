package dtmocks

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal/dalmocks"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db"
	"github.com/strongo/db/mockdb"
)

func SetupMocks(c context.Context) {
	mockDB := mockdb.NewMockDB(nil, nil)

	dal.Transfer = dalmocks.NewTransferDalMock(mockDB)
	dal.User = dalmocks.NewUserDalMock()
	dal.Contact = dalmocks.NewContactDalMock()

	if err := mockDB.UpdateMulti(c, []db.EntityHolder{
		&models.AppUser{IntegerID: db.IntegerID{ID: 1}, AppUserEntity: &models.AppUserEntity{ContactDetails: models.ContactDetails{FirstName: "Alfred", LastName: "Alpha"}}},
		&models.AppUser{IntegerID: db.IntegerID{ID: 3}, AppUserEntity: &models.AppUserEntity{ContactDetails: models.ContactDetails{FirstName: "Ben", LastName: "Bravo"}}},
		&models.AppUser{IntegerID: db.IntegerID{ID: 5}, AppUserEntity: &models.AppUserEntity{ContactDetails: models.ContactDetails{FirstName: "Charles", LastName: "Cain"}}},
	}); err != nil {
		panic(err)
	}

	if err := mockDB.UpdateMulti(c, []db.EntityHolder{
		&models.Contact{
			IntegerID: db.NewIntID(2),
			ContactEntity: &models.ContactEntity{
				Status:             models.STATUS_ACTIVE,
				UserID:             1,
				CounterpartyUserID: 3,
				ContactDetails:     models.ContactDetails{Nickname: "Bono"}},
		},
		&models.Contact{
			IntegerID: db.NewIntID(4),
			ContactEntity: &models.ContactEntity{
				Status:             models.STATUS_ACTIVE,
				UserID:             1,
				CounterpartyUserID: 5,
				ContactDetails:     models.ContactDetails{Nickname: "Carly"}},
		},
		&models.Contact{IntegerID: db.NewIntID(6), ContactEntity: &models.ContactEntity{
			Status: models.STATUS_ACTIVE, UserID: 1, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Den"}}},
		&models.Contact{IntegerID: db.NewIntID(62), ContactEntity: &models.ContactEntity{
			Status: models.STATUS_ACTIVE, UserID: 1, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Den 2"}}},
		&models.Contact{IntegerID: db.NewIntID(63), ContactEntity: &models.ContactEntity{
			Status: models.STATUS_ACTIVE, UserID: 1, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Den 3"}}},
		&models.Contact{IntegerID: db.NewIntID(8), ContactEntity: &models.ContactEntity{
			Status: models.STATUS_ACTIVE, UserID: 3, CounterpartyUserID: 1, ContactDetails: models.ContactDetails{Nickname: "Eagle"}}},
		&models.Contact{IntegerID: db.NewIntID(10), ContactEntity: &models.ContactEntity{
			Status: models.STATUS_ACTIVE, UserID: 5, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Ford"}}},
		&models.Contact{IntegerID: db.NewIntID(12), ContactEntity: &models.ContactEntity{
			Status: models.STATUS_ACTIVE, UserID: 5, CounterpartyUserID: 0, ContactDetails: models.ContactDetails{Nickname: "Gina"}}},
	}); err != nil {
		panic(err)
	}

	dal.DB = mockDB
}
