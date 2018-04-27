package gaedal

import (
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

type UserGaClientDalGae struct {
}

func NewUserGaClientDalGae() UserGaClientDalGae {
	return UserGaClientDalGae{}
}

func (UserGaClientDalGae) SaveGaClient(c context.Context, gaClientId, userAgent, ipAddress string) (gaClient models.GaClient, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		var entity models.GaClientEntity
		key := gaedb.NewKey(c, models.GaClientKind, gaClientId, 0, nil)
		err := gaedb.Get(c, key, &entity)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return errors.Wrap(err, "Failed to get UserGaClient by ID")
		}
		if entity.UserAgent != userAgent || entity.IpAddress != ipAddress {
			entity.UserAgent = userAgent
			entity.IpAddress = ipAddress
			entity.Created = time.Now()
			if _, err = gaedb.Put(c, key, entity); err != nil {
				err = errors.Wrap(err, "Failed to save UserGaClient")
				return err
			}
		}
		return nil
	}, nil)
	return
}
