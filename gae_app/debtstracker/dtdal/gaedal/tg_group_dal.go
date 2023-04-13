package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
)

type tgGroupDalGae struct {
}

func newTgGroupDalGae() tgGroupDalGae {
	return tgGroupDalGae{}
}

func (tgGroupDalGae) GetTgGroupByID(c context.Context, tx dal.ReadSession, id int64) (tgGroup models.TgGroup, err error) {
	tgGroup = models.NewTgGroup(id, nil)
	if tx == nil {
		if tx, err = facade.GetDatabase(c); err != nil {
			return
		}
	}
	return tgGroup, tx.Get(c, tgGroup.Record)
}

func (tgGroupDalGae) SaveTgGroup(c context.Context, tx dal.ReadwriteTransaction, tgGroup models.TgGroup) (err error) {
	return tx.Set(c, tgGroup.Record)
}
