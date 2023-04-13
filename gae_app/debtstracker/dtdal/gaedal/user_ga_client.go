package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/v2/datastore"
)

type UserGaClientDalGae struct {
}

func NewUserGaClientDalGae() UserGaClientDalGae {
	return UserGaClientDalGae{}
}

func (UserGaClientDalGae) SaveGaClient(c context.Context, gaClientId, userAgent, ipAddress string) (gaClient models.GaClient, err error) {
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		var entity models.GaClientEntity
		key := gaedb.NewKey(c, models.GaClientKind, gaClientId, 0, nil)
		err := gaedb.Get(c, key, &entity)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return fmt.Errorf("failed to get UserGaClient by ID: %w", err)
		}
		if entity.UserAgent != userAgent || entity.IpAddress != ipAddress {
			entity.UserAgent = userAgent
			entity.IpAddress = ipAddress
			entity.Created = time.Now()
			if _, err = gaedb.Put(c, key, entity); err != nil {
				err = fmt.Errorf("failed to save UserGaClient: %w", err)
				return err
			}
		}
		return nil
	}, nil)
	return
}
