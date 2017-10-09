package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/app/db"
)

func newUserGooglePlusKey(c context.Context, id string) *datastore.Key {
	return gaedb.NewKey(c, models.UserGooglePlusKind, id, 0, nil)
}

type UserGooglePlusDalGae struct {
}

func NewUserGooglePlusDalGae() UserGooglePlusDalGae {
	return UserGooglePlusDalGae{}
}

func (_ UserGooglePlusDalGae) GetUserGooglePlusByID(c context.Context, id string) (userGooglePlus models.UserGooglePlus, err error) {
	var userGooglePlusEntity models.UserGooglePlusEntity
	if err = gaedb.Get(c, newUserGooglePlusKey(c, id), &userGooglePlusEntity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.NewErrNotFoundByStrID(models.UserGooglePlusKind, id, err)
		}
		return
	}
	userGooglePlus = models.UserGooglePlus{ID: id, UserGooglePlusEntity: &userGooglePlusEntity}
	return
}

func (_ UserGooglePlusDalGae) SaveUserGooglePlusByID(c context.Context, userGooglePlus models.UserGooglePlus) (err error) {
	if _, err = gaedb.Put(c, newUserGooglePlusKey(c, userGooglePlus.ID), userGooglePlus.UserGooglePlusEntity); err != nil {
		return
	}
	return
}
