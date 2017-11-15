package models

import "github.com/strongo/app/db"

const GroupMemberKind = "GroupMember"

type GroupMember struct {
	db.IntegerID
	*GroupMemberEntity
}

var _ db.EntityHolder = (*GroupMember)(nil)

func (GroupMember) Kind() string {
	return GroupMemberKind
}

func (gm GroupMember) Entity() interface{} {
	return gm.GroupMemberEntity
}

func (gm *GroupMember) SetEntity(entity interface{}) {
	gm.GroupMemberEntity = entity.(*GroupMemberEntity)
}

type GroupMemberEntity struct {
	GroupID    int64
	UserID     int64
	ContactIDs []int64
	Name       string `datastore:",noindex"`
}
