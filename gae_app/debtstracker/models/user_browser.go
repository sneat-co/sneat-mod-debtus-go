package models

import (
	"time"

	"github.com/strongo/db"
)

const UserBrowserKind = "UserBrowser"

type UserBrowserEntity struct {
	UserID      int64
	UserAgent   string
	LastUpdated time.Time `datastore:",noindex"`
}

type UserBrowser struct {
	db.IntegerID
	*UserBrowserEntity
}

var _ db.EntityHolder = (*UserBrowser)(nil)

func (UserBrowser) Kind() string {
	return UserBrowserKind
}

func (ub UserBrowser) Entity() interface{} {
	return ub.UserBrowserEntity
}

func (UserBrowser) NewEntity() interface{} {
	return new(UserBrowserEntity)
}

func (ub *UserBrowser) SetEntity(entity interface{}) {
	if entity == nil {
		ub.UserBrowserEntity = nil
	} else {
		ub.UserBrowserEntity = entity.(*UserBrowserEntity)
	}
}
