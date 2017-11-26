package models

import (
	"time"

	"github.com/strongo/db"
)

const LoginPinKind = "LoginPin"

type LoginPin struct {
	db.IntegerID
	*LoginPinEntity
}

var _ db.EntityHolder = (*LoginPin)(nil)

func (LoginPin) Kind() string {
	return LoginPinKind
}

func (loginPin LoginPin) Entity() interface{} {
	return loginPin.LoginPinEntity
}

func (LoginPin) NewEntity() interface{} {
	return new(LoginPinEntity)
}

func (loginPin *LoginPin) SetEntity(entity interface{}) {
	if entity == nil {
		loginPin.LoginPinEntity = nil
	} else {
		loginPin.LoginPinEntity = entity.(*LoginPinEntity)
	}

}

type LoginPinEntity struct {
	Channel    string `datastore:",noindex"`
	GaClientID string `datastore:",noindex"`
	Created    time.Time
	Pinned     time.Time `datastore:",noindex"`
	SignedIn   time.Time `datastore:",noindex"`
	UserID     int64     `datastore:",noindex"`
	Code       int32     `datastore:",noindex"`
}

func (entity *LoginPinEntity) IsActive(channel string) bool {
	return entity.SignedIn.IsZero() && entity.Channel == channel
}
