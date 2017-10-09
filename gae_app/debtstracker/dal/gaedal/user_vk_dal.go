package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func NewUserVkKey(c context.Context, vkUserID int64) *datastore.Key {
	return gaedb.NewKey(c, models.UserVkKind, "", vkUserID, nil)
}

type UserVkDalGae struct {
}

func NewUserVkDalGae() UserVkDalGae {
	return UserVkDalGae{}
}

func (_ UserVkDalGae) GetUserVkByID(c context.Context, vkUserID int64) (vkUser models.UserVk, err error) {
	vkUserKey := NewUserVkKey(c, vkUserID)
	var vkUserEntity models.UserVkEntity
	if err = gaedb.Get(c, vkUserKey, &vkUserEntity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.NewErrNotFoundByIntID(models.UserVkKind, vkUserID, nil)
		}
		return
	}
	vkUser = models.UserVk{ID: vkUserID, UserVkEntity: &vkUserEntity}
	return
}

func (_ UserVkDalGae) SaveUserVk(c context.Context, userVk models.UserVk) (err error) {
	k := NewUserVkKey(c, userVk.ID)
	_, err = gaedb.Put(c, k, userVk)
	return
}
