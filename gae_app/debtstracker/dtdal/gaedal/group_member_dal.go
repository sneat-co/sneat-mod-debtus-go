package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
)

type GroupMemberDalGae struct {
}

func NewGroupMemberDalGae() GroupMemberDalGae {
	return GroupMemberDalGae{}
}

func (GroupMemberDalGae) CreateGroupMember(c context.Context, tx dal.ReadwriteTransaction, groupMemberData *models.GroupMemberData) (groupMember models.GroupMember, err error) {
	key := models.NewGroupMemberIncompleteKey()
	groupMember.Record = dal.NewRecordWithData(key, groupMemberData)
	if err = tx.Insert(c, groupMember.Record); err != nil {
		return
	}
	groupMember.ID = groupMember.Record.Key().ID.(int64)
	return
}

func (GroupMemberDalGae) GetGroupMemberByID(c context.Context, tx dal.ReadSession, groupMemberID int64) (groupMember models.GroupMember, err error) {
	groupMember = models.NewGroupMember(groupMemberID, nil)
	if tx == nil {
		if tx, err = facade.GetDatabase(c); err != nil {
			return
		}
	}
	return groupMember, tx.Get(c, groupMember.Record)
}
