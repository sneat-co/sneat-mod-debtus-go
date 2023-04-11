package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/v2/datastore"
)

func NewPasswordResetKey(c context.Context, id int64) *datastore.Key {
	return datastore.NewKey(c, models.PasswordResetKind, "", id, nil)
}

func NewPasswordResetIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.PasswordResetKind, nil)
}

func NewPasswordResetDalGae() PasswordResetDalGae {
	return PasswordResetDalGae{}
}

type PasswordResetDalGae struct {
}

var _ dtdal.PasswordResetDal = (*PasswordResetDalGae)(nil)

func (PasswordResetDalGae) GetPasswordResetByID(c context.Context, id int64) (passwordReset models.PasswordReset, err error) {
	key := NewPasswordResetKey(c, id)
	passwordReset.ID = id
	passwordReset.PasswordResetEntity = new(models.PasswordResetEntity)
	if err = gaedb.Get(c, key, passwordReset.PasswordResetEntity); err == datastore.ErrNoSuchEntity {
		err = db.NewErrNotFoundByIntID(models.PasswordResetKind, id, err)
	}
	return
}

func (PasswordResetDalGae) CreatePasswordResetByID(c context.Context, entity *models.PasswordResetEntity) (passwordReset models.PasswordReset, err error) {
	key := NewPasswordResetIncompleteKey(c)
	if key, err = gaedb.Put(c, key, entity); err != nil {
		return
	}
	passwordReset.ID = key.IntID()
	passwordReset.PasswordResetEntity = entity
	return
}

func (PasswordResetDalGae) SavePasswordResetByID(c context.Context, entity models.PasswordReset) (err error) {
	key := NewPasswordResetKey(c, entity.ID)
	_, err = gaedb.Put(c, key, entity.PasswordResetEntity)
	return
}
