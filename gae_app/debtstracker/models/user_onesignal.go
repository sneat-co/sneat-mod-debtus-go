package models

import (
	"github.com/strongo/app/db"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
)

const UserOneSignalKind = "UserOneSignal"

type UserOneSignalEntity struct {
	UserID  int64
	Created time.Time
}

type UserOneSignal struct {
	db.NoIntID
	ID string
	UserOneSignalEntity
}

func NewUserOneSignalKey(c context.Context, oneSignalUserID string) *datastore.Key {
	return datastore.NewKey(c, UserOneSignalKind, oneSignalUserID, 0, nil)
}
