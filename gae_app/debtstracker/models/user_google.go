package models

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app/user"
)

const UserGoogleCollection = "UserAccount"
const UserFacebookCollection = "UserFb"

var _ user.AccountRecord = (*UserAccount)(nil)

// UserAccount - TODO: consider migrating to https://github.com/dal-go/dalgo4auth
type UserAccount struct { // TODO: Move out to library?
	record.WithID[string]
	data *user.AccountDataBase
}

func (ua UserAccount) Key() user.Account {
	return ua.data.Account
}

func (ua UserAccount) Data() user.AccountData {
	return ua.data
}

func (ua UserAccount) DataStruct() *user.AccountDataBase {
	return ua.data
}

func NewUserAccount(id string) UserAccount {
	key := dal.NewKeyWithID(UserGoogleCollection, id)
	data := new(user.AccountDataBase)
	return UserAccount{
		WithID: record.WithID[string]{
			ID:     id,
			Key:    key,
			Record: dal.NewRecordWithData(key, data),
		},
		data: data,
	}
}
