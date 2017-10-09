package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
	"github.com/strongo/app/gaedb"
)

type UserOneSignalDalGae struct {
}

func NewUserOneSignalDalGae() UserOneSignalDalGae {
	return UserOneSignalDalGae{}
}


func (_ UserOneSignalDalGae) SaveUserOneSignal(c context.Context, userID int64, oneSignalUserID string) (userOneSignal models.UserOneSignal, err error) {
	key := models.NewUserOneSignalKey(c, oneSignalUserID)
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
	userOneSignal = models.UserOneSignal{ID: oneSignalUserID, UserOneSignalEntity: entity}
	return
}
