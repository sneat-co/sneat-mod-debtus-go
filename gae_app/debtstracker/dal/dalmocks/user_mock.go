package dalmocks

import (
	"strconv"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db"
)

type UserDalMock struct {
	LastUserID int64
	Users      map[int64]*models.AppUserEntity
}

func (mock *UserDalMock) DelayUpdateUserWithContact(c context.Context, userID, contactID int64) error {
	panic("not implemented yet")
}

func NewUserDalMock() *UserDalMock {
	return &UserDalMock{
		Users: make(map[int64]*models.AppUserEntity),
	}
}

func (mock *UserDalMock) SetLastCurrency(c context.Context, userID int64, currency models.Currency) error {
	panic("Not implemented yet")
}

func (mock *UserDalMock) GetUserByID(c context.Context, userID int64) (appUser models.AppUser, err error) {
	if entity, ok := mock.Users[userID]; ok {
		return models.AppUser{IntegerID: db.NewIntID(userID), AppUserEntity: entity}, nil
	}
	return models.AppUser{IntegerID: db.NewIntID(userID)}, db.NewErrNotFoundByIntID(models.AppUserKind, userID, nil)
}

func (mock *UserDalMock) GetUserByStrID(c context.Context, userID string) (user models.AppUser, err error) {
	if user.ID, err = strconv.ParseInt(userID, 10, 64); err != nil {
		return
	}
	return mock.GetUserByID(c, user.ID)
}

func (mock *UserDalMock) GetUsersByIDs(c context.Context, userIDs []int64) (users []models.AppUser, err error) {
	users = make([]models.AppUser, len(userIDs))
	for i, userID := range userIDs {
		if users[i], err = mock.GetUserByID(c, userID); err != nil {
			return
		}
	}
	return
}

func (mock *UserDalMock) GetUserByEmail(c context.Context, email string) (models.AppUser, error) {
	panic("Not implemented yet")
}

func (mock *UserDalMock) CreateUser(c context.Context, userEntity *models.AppUserEntity) (models.AppUser, error) {
	panic("Not implemented yet")
}

func (mock *UserDalMock) GetUserByVkUserID(c context.Context, vkUserID int64) (models.AppUser, error) {
	panic("Not implemented yet")
}
func (mock *UserDalMock) CreateAnonymousUser(c context.Context) (models.AppUser, error) {
	panic("Not implemented yet")
}
func (mock *UserDalMock) SaveUser(c context.Context, user models.AppUser) error {
	return nil
}
func (mock *UserDalMock) DelaySetUserPreferredLocale(c context.Context, delay time.Duration, userID int64, localeCode5 string) error {
	return nil
}
func (mock *UserDalMock) DelayUpdateUserHasDueTransfers(c context.Context, userID int64) error {
	return nil
}
func (mock *UserDalMock) DelayUpdateUserWithBill(c context.Context, groupID, billID string) error {
	panic(NOT_IMPLEMENTED_YET)
}
