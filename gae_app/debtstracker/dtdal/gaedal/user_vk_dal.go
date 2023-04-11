package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/v2/datastore"
)

func NewUserVkKey(c context.Context, vkUserID int64) *datastore.Key {
	return gaedb.NewKey(c, models.UserVkKind, "", vkUserID, nil)
}

type UserVkDalGae struct {
}

func NewUserVkDalGae() UserVkDalGae {
	return UserVkDalGae{}
}

func (UserVkDalGae) GetUserVkByID(c context.Context, vkUserID int64) (vkUser models.UserVk, err error) {
	vkUserKey := NewUserVkKey(c, vkUserID)
	var vkUserEntity models.UserVkEntity
	if err = gaedb.Get(c, vkUserKey, &vkUserEntity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.NewErrNotFoundByIntID(models.UserVkKind, vkUserID, nil)
		}
		return
	}
	vkUser = models.UserVk{IntegerID: db.NewIntID(vkUserID), UserVkEntity: &vkUserEntity}
	return
}

func (UserVkDalGae) SaveUserVk(c context.Context, userVk models.UserVk) (err error) {
	k := NewUserVkKey(c, userVk.ID)
	_, err = gaedb.Put(c, k, userVk)
	return
}
