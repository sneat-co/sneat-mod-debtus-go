package models

import "github.com/strongo/app/db"

const GroupMemberKind = "GroupMember"

type GroupMember struct {
	db.NoStrID
	ID int64
	*GroupMemberEntity
}

type GroupMemberEntity struct {
	GroupID    int64
	UserID     int64
	ContactIDs []int64
	Name       string `datastore:",noindex"`
}
