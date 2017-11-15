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
	db.IntegerID
	*UserVkEntity
}

var _ db.EntityHolder = (*UserVk)(nil)

func (UserVk) Kind() string {
	return  UserVkKind
}

func (u UserVk) Entity() interface{} {
	return u.UserVkEntity
}

func (u *UserVk) SetEntity(entity interface{}) {
	u.UserVkEntity = entity.(*UserVkEntity)
}

