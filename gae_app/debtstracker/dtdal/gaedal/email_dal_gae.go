package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
)

type EmailDalGae struct {
}

func NewEmailDalGae() EmailDalGae {
	return EmailDalGae{}
}

var _ dtdal.EmailDal = (*EmailDalGae)(nil)

func (EmailDalGae) InsertEmail(c context.Context, tx dal.ReadwriteTransaction, data *models.EmailData) (email models.Email, err error) {
	key := dal.NewKey(models.EmailKind)
	email.Record = dal.NewRecordWithData(key, data)
	if err = tx.Insert(c, email.Record); err != nil {
		return
	}
	email.ID = email.Record.Key().ID.(int64)
	email.Data = data
	return
}

func (EmailDalGae) UpdateEmail(c context.Context, tx dal.ReadwriteTransaction, email models.Email) (err error) {
	return tx.Set(c, email.Record)
}

func (EmailDalGae) GetEmailByID(c context.Context, tx dal.ReadSession, id int64) (email models.Email, err error) {
	email = models.NewEmail(id, nil)
	return email, tx.Get(c, email.Record)
}
