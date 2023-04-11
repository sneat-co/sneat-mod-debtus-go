package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/v2/datastore"
)

type EmailDalGae struct {
}

func NewEmailDalGae() EmailDalGae {
	return EmailDalGae{}
}

var _ dtdal.EmailDal = (*EmailDalGae)(nil)

func NewEmailKey(c context.Context, id int64) *datastore.Key {
	return datastore.NewKey(c, models.EmailKind, "", id, nil)
}

func NewEmailIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.EmailKind, nil)
}

func (EmailDalGae) InsertEmail(c context.Context, entity *models.EmailEntity) (email models.Email, err error) {
	key := NewEmailIncompleteKey(c)
	if key, err = gaedb.Put(c, key, entity); err != nil {
		return
	}
	email.ID = key.IntID()
	email.EmailEntity = entity
	return
}

func (EmailDalGae) UpdateEmail(c context.Context, email models.Email) (err error) {
	if email.ID == 0 {
		return errors.New("UpdateEmail(email.ID == 0)")
	}
	if email.EmailEntity == nil {
		return errors.New("UpdateEmail(email.EmailEntity == nil)")
	}
	_, err = gaedb.Put(c, NewEmailKey(c, email.ID), email.EmailEntity)
	return
}

func (EmailDalGae) GetEmailByID(c context.Context, id int64) (email models.Email, err error) {
	email.ID = id
	emailEntity := new(models.EmailEntity)
	if err = gaedb.Get(c, NewEmailKey(c, id), emailEntity); err != nil {
		return
	}
	email.EmailEntity = emailEntity
	return
}
