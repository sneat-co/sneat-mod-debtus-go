package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

func NewUserEmailKey(c context.Context, email string) *datastore.Key {
	return datastore.NewKey(c, models.UserEmailKind, models.GetEmailID(email), 0, nil)
}

type UserEmailGaeDal struct {
}

func NewUserEmailGaeDal() UserEmailGaeDal {
	return UserEmailGaeDal{}
}

func (UserEmailGaeDal) GetUserEmailByID(c context.Context, email string) (userEmail models.UserEmail, err error) {
	userEmail.UserEmailEntity = new(models.UserEmailEntity)
	key := NewUserEmailKey(c, email)
	userEmail.ID = key.StringID()
	if err = gaedb.Get(c, key, userEmail.UserEmailEntity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.ErrRecordNotFound
		}
		return
	}
	return
}

func (UserEmailGaeDal) SaveUserEmail(c context.Context, userEmail models.UserEmail) (err error) {
	_, err = gaedb.Put(c, NewUserEmailKey(c, userEmail.ID), userEmail.UserEmailEntity)
	return
}
