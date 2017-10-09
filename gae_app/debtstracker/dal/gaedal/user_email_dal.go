package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"golang.org/x/net/context"
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

func (_ UserEmailGaeDal) GetUserEmailByID(c context.Context, email string) (userEmail models.UserEmail, err error) {
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

func (_ UserEmailGaeDal) SaveUserEmail(c context.Context, userEmail models.UserEmail) (err error) {
	_, err = gaedb.Put(c, NewUserEmailKey(c, userEmail.ID), userEmail.UserEmailEntity)
	return
}
