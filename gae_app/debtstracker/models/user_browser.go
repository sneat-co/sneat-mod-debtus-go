package models

import (
	"github.com/strongo/app/db"
	"time"
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

func (ub *UserBrowser) SetEntity(entity interface{}) {
	ub.UserBrowserEntity = entity.(*UserBrowserEntity)
}

