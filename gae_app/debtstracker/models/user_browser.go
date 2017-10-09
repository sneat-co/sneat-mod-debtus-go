package models

import (
	"time"
	"github.com/strongo/app/db"
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
