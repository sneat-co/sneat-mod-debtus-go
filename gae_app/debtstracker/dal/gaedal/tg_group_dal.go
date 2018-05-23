package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

type tgGroupDalGae struct {
}

func newTgGroupDalGae() tgGroupDalGae {
	return tgGroupDalGae{}
}

func newTgGroupKey(c context.Context, id int64) *datastore.Key {
	return datastore.NewKey(c, models.TgGroupKind, "", id, nil)
}

func (tgGroupDalGae) GetTgGroupByID(c context.Context, id int64) (tgGroup models.TgGroup, err error) {
	tgGroup.TgGroupEntity = new(models.TgGroupEntity)
	err = get(c, newTgGroupKey(c, id), &tgGroup)
	return
}

func (tgGroupDalGae) SaveTgGroup(c context.Context, tgGroup models.TgGroup) (err error) {
	if _, err = gaedb.Put(c, newTgGroupKey(c, tgGroup.ID), tgGroup.TgGroupEntity); err != nil {
		return
	}
	return
}

func get(c context.Context, key *datastore.Key, entityHolder db.EntityHolder) (err error) {
	kind := entityHolder.Kind()
	if err = gaedb.Get(c, key, entityHolder.Entity()); err != nil {
		entityHolder.SetEntity(nil)
		if err == datastore.ErrNoSuchEntity {
			if intID := entityHolder.IntID(); intID != 0 {
				entityHolder.SetIntID(0)
				err = db.NewErrNotFoundByIntID(kind, intID, err)
			} else if strID := entityHolder.StrID(); strID != "" {
				err = db.NewErrNotFoundByStrID(kind, strID, err)
			} else {
				err = db.ErrRecordNotFound
			}
		} else {
			err = errors.Wrapf(err, "failed to get entity by key=%v", key)
		}
	}
	return
}
