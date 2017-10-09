package models

import (
	"github.com/strongo/app/db"
	"github.com/strongo/app/user"
)

const (
	UserVkKind = "UserVk"
)

type UserVkEntity struct {
	user.OwnedByUser
	FirstName  string
	LastName   string
	ScreenName string
	Nickname   string
	//FriendIDs []int64 `datastore:",noindex"`
}

type UserVk struct {
	db.NoStrID
	ID int64
	*UserVkEntity
}
