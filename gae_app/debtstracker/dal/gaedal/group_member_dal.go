package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

//var _ dal.GroupMemberDal = (*GroupMemberDalGae)(nil)

func NewGroupMemberKey(c context.Context, groupMemberID int64) *datastore.Key {
	if groupMemberID == 0 {
		panic("groupMemberID == 0")
	}
	return gaedb.NewKey(c, models.GroupMemberKind, "", groupMemberID, nil)
}

func NewGroupMemberIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.GroupMemberKind, nil)
}

type GroupMemberDalGae struct {
}

func NewGroupMemberDalGae() GroupMemberDalGae {
	return GroupMemberDalGae{}
}

func (GroupMemberDalGae) CreateGroupMember(c context.Context, groupMemberEntity *models.GroupMemberEntity) (groupMember models.GroupMember, err error) {
	key := NewGroupMemberIncompleteKey(c)
	key, err = gaedb.Put(c, key, groupMemberEntity)
	groupMember = models.GroupMember{IntegerID: db.NewIntID(key.IntID()), GroupMemberEntity: groupMemberEntity}
	return
}

func (GroupMemberDalGae) GetGroupMemberByID(c context.Context, groupMemberID int64) (groupMember models.GroupMember, err error) {
	groupMemberEntity := new(models.GroupMemberEntity)
	err = gaedb.Get(c, NewGroupMemberKey(c, groupMemberID), groupMemberEntity)
	groupMember = models.GroupMember{IntegerID: db.NewIntID(groupMemberID), GroupMemberEntity: groupMemberEntity}
	return
}
