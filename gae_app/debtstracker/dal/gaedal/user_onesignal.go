package gaedal

import (
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

type UserOneSignalDalGae struct {
}

func NewUserOneSignalDalGae() UserOneSignalDalGae {
	return UserOneSignalDalGae{}
}

func (userOneSignalDalGae UserOneSignalDalGae) SaveUserOneSignal(c context.Context, userID int64, oneSignalUserID string) (userOneSignal models.UserOneSignal, err error) {
	key := userOneSignalDalGae.NewUserOneSignalKey(c, oneSignalUserID)
	var entity models.UserOneSignalEntity
	// Save if no entity or AppUserIntID changed
	if err = gaedb.Get(c, key, &entity); err == datastore.ErrNoSuchEntity || entity.UserID != userID {
		entity = models.UserOneSignalEntity{UserID: userID, Created: time.Now()}
		if _, err = gaedb.Put(c, key, &entity); err != nil {
			return
		}
	} else if err != nil {
		return
	}
	userOneSignal = models.UserOneSignal{StringID: db.StringID{ID: oneSignalUserID}, UserOneSignalEntity: &entity}
	return
}

func (UserOneSignalDalGae) NewUserOneSignalKey(c context.Context, oneSignalUserID string) *datastore.Key {
	return datastore.NewKey(c, models.UserOneSignalKind, oneSignalUserID, 0, nil)
}
