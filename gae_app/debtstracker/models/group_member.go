package models

import (
	"github.com/strongo/dalgo/record"
)

const GroupMemberKind = "GroupMember"

type GroupMember struct {
	record.WithID[int]
	*GroupMemberEntity
}

//var _ db.EntityHolder = (*GroupMember)(nil)

func (GroupMember) Kind() string {
	return GroupMemberKind
}

func (gm GroupMember) Entity() interface{} {
	return gm.GroupMemberEntity
}

func (GroupMember) NewEntity() interface{} {
	return new(GroupMemberEntity)
}

func (gm *GroupMember) SetEntity(entity interface{}) {
	if entity == nil {
		gm.GroupMemberEntity = nil
	} else {
		gm.GroupMemberEntity = entity.(*GroupMemberEntity)
	}
}

type GroupMemberEntity struct {
	GroupID    int64
	UserID     int64
	ContactIDs []int64
	Name       string `datastore:",noindex"`
}
