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
	db.NoStrID
	ID int64
	*UserBrowserEntity
}
