package models

import (
	"github.com/strongo/app/db"
	"time"
)

const UserOneSignalKind = "UserOneSignal"

type UserOneSignalEntity struct {
	UserID  int64
	Created time.Time
}

type UserOneSignal struct {
	db.StringID
	*UserOneSignalEntity
}

var _ db.EntityHolder = (*UserOneSignal)(nil)

func (UserOneSignal) Kind() string {
	return UserOneSignalKind
}

func (userOneSignal UserOneSignal) Entity() interface{} {
	return userOneSignal.UserOneSignalEntity
}

func (userOneSignal *UserOneSignal) SetEntity(entity interface{}) {
	userOneSignal.UserOneSignalEntity = entity.(*UserOneSignalEntity)
}
