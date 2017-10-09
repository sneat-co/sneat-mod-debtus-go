package models

import (
	"time"
	"github.com/strongo/app/db"
)

const LoginPinKind = "LoginPin"

type LoginPin struct {
	ID int64
	db.NoStrID
	*LoginPinEntity
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