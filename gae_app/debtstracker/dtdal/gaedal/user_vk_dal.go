package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/dal-go/dalgo/dal"
)

func NewUserVkKey(vkUserID int64) *dal.Key {
	return dal.NewKeyWithID(models.UserVkKind, vkUserID)
}

type UserVkDalGae struct {
}

func NewUserVkDalGae() UserVkDalGae {
	return UserVkDalGae{}
}

func (UserVkDalGae) GetUserVkByID(c context.Context, vkUserID int64) (vkUser models.UserVk, err error) {
	//vkUserKey := NewUserVkKey(c, vkUserID)
	//var vkUserEntity models.UserVkEntity
	//if err = gaedb.Get(c, vkUserKey, &vkUserEntity); err != nil {
	//	if err == datastore.ErrNoSuchEntity {
	//		err = db.NewErrNotFoundByIntID(models.UserVkKind, vkUserID, nil)
	//	}
	//	return
	//}
	//vkUser = models.UserVk{IntegerID: db.NewIntID(vkUserID), UserVkEntity: &vkUserEntity}
	return vkUser, errors.New("not implemented")
}

func (UserVkDalGae) SaveUserVk(c context.Context, userVk models.UserVk) (err error) {
	//k := NewUserVkKey(c, userVk.ID)
	//_, err = gaedb.Put(c, k, userVk)
	return errors.New("not implemented")
}
