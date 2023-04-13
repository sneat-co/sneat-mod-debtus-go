package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
)

func NewPasswordResetDalGae() PasswordResetDalGae {
	return PasswordResetDalGae{}
}

type PasswordResetDalGae struct {
}

var _ dtdal.PasswordResetDal = (*PasswordResetDalGae)(nil)

func (PasswordResetDalGae) GetPasswordResetByID(c context.Context, tx dal.ReadSession, id int) (passwordReset models.PasswordReset, err error) {
	passwordReset = models.NewPasswordReset(id, nil)
	if tx == nil {
		if tx, err = facade.GetDatabase(c); err != nil {
			return
		}
	}
	return passwordReset, tx.Get(c, passwordReset.Record)
}

func (PasswordResetDalGae) CreatePasswordResetByID(c context.Context, tx dal.ReadwriteTransaction, entity *models.PasswordResetData) (passwordReset models.PasswordReset, err error) {
	passwordReset = models.NewPasswordReset(0, entity)
	if err = tx.Insert(c, passwordReset.Record); err != nil {
		return
	}
	passwordReset.ID = passwordReset.Key.ID.(int)
	return
}

func (PasswordResetDalGae) SavePasswordResetByID(c context.Context, tx dal.ReadwriteTransaction, passwordReset models.PasswordReset) (err error) {
	return tx.Set(c, passwordReset.Record)
}
