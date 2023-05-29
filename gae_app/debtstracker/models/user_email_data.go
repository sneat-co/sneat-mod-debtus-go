package models

import (
	"encoding/base64"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app/user"
)

func NewUserEmailKey(email string) *dal.Key {
	return dal.NewKeyWithID(UserEmailKind, GetEmailID(email))
}

func NewUserEmail(email string, data *UserEmailData) UserEmail {
	id := GetEmailID(email)
	if data == nil {
		data = new(UserEmailData)
	}
	return UserEmail{
		WithID: record.NewWithID(id, NewUserEmailKey(email), data),
		Data:   data,
	}
}

var _ user.AccountData = (*UserEmailData)(nil)

func (entity *UserEmailData) GetNames() user.Names {
	return entity.Names
}

func (entity *UserEmailData) ConfirmationPin() string {
	pin := base64.RawURLEncoding.EncodeToString(entity.PasswordBcryptHash)
	//if len(pin) > 20 {
	//	pin = pin[:20]
	//}
	return pin
}

func (entity *UserEmailData) IsEmailConfirmed() bool {
	return entity.IsConfirmed
}
