package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
)

//func newUserGooglePlusKey(id string) *dal.Key {
//	return dal.NewKeyWithID(models.UserGooglePlusKind, id)
//}

type UserGooglePlusDalGae struct {
}

func NewUserGooglePlusDalGae() UserGooglePlusDalGae {
	return UserGooglePlusDalGae{}
}

func (UserGooglePlusDalGae) GetUserGooglePlusByID(c context.Context, id string) (userGooglePlus models.UserGooglePlus, err error) {
	//var userGooglePlusEntity models.UserGooglePlusEntity
	//if err = gaedb.Get(c, newUserGooglePlusKey(id), &userGooglePlusEntity); err != nil {
	//	if err == datastore.ErrNoSuchEntity {
	//		err = db.NewErrNotFoundByStrID(models.UserGooglePlusKind, id, err)
	//	}
	//	return
	//}
	//userGooglePlus = models.UserGooglePlus{StringID: db.StringID{ID: id}, UserGooglePlusEntity: &userGooglePlusEntity}
	err = errors.New("not implemented")
	return
}

func (UserGooglePlusDalGae) SaveUserGooglePlusByID(c context.Context, userGooglePlus models.UserGooglePlus) (err error) {
	//if _, err = gaedb.Put(c, newUserGooglePlusKey(userGooglePlus.ID), userGooglePlus.UserGooglePlusEntity); err != nil {
	//	return
	//}
	return errors.New("not implemented")
}
