package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/db/gaedb"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type EmailDalGae struct {
}

func NewEmailDalGae() EmailDalGae {
	return EmailDalGae{}
}

var _ dal.EmailDal = (*EmailDalGae)(nil)

func NewEmailKey(c context.Context, id int64) *datastore.Key {
	return datastore.NewKey(c, models.EmailKind, "", id, nil)
}

func NewEmailIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.EmailKind, nil)
}

func (_ EmailDalGae) InsertEmail(c context.Context, entity *models.EmailEntity) (email models.Email, err error) {
	key := NewEmailIncompleteKey(c)
	if key, err = gaedb.Put(c, key, entity); err != nil {
		return
	}
	email.ID = key.IntID()
	email.EmailEntity = entity
	return
}

func (_ EmailDalGae) UpdateEmail(c context.Context, email models.Email) (err error) {
	if email.ID == 0 {
		return errors.New("UpdateEmail(email.ID == 0)")
	}
	if email.EmailEntity == nil {
		return errors.New("UpdateEmail(email.EmailEntity == nil)")
	}
	_, err = gaedb.Put(c, NewEmailKey(c, email.ID), email.EmailEntity)
	return
}

func (_ EmailDalGae) GetEmailByID(c context.Context, id int64) (email models.Email, err error) {
	email.ID = id
	emailEntity := new(models.EmailEntity)
	if err = gaedb.Get(c, NewEmailKey(c, id), emailEntity); err != nil {
		return
	}
	email.EmailEntity = emailEntity
	return
}
