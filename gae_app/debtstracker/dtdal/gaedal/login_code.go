package gaedal

import (
	"fmt"
	"math/rand"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
)

type LoginCodeDalGae struct {
}

func NewLoginCodeDalGae() LoginCodeDalGae {
	return LoginCodeDalGae{}
}

func NewLoginCodeKey(c context.Context, code int32) *datastore.Key {
	return gaedb.NewKey(c, models.LoginCodeKind, "", int64(code), nil)
}

func (LoginCodeDalGae) NewLoginCode(c context.Context, userID int64) (int32, error) {
	var code int32
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 1; i < 20; i++ {
		code = random.Int31n(99999) + 1
		key := NewLoginCodeKey(c, code)
		if err := gaedb.Get(c, key, &models.LoginCodeEntity{}); err == datastore.ErrNoSuchEntity {
			var created bool
			err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
				var entity models.LoginCodeEntity
				if err := gaedb.Get(c, key, &entity); err == datastore.ErrNoSuchEntity || entity.Created.Add(time.Hour).Before(time.Now()) {
					entity = models.LoginCodeEntity{
						Created: time.Now(),
						UserID:  userID,
					}
					if _, err = gaedb.Put(c, key, &entity); err != nil {
						log.Errorf(c, fmt.Errorf("failed to save %v: %w", key, err).Error())
					}
					created = true
					return nil
				} else if err != nil {
					return fmt.Errorf("failed to get entity within transaction: %w", err)
				} else {
					log.Warningf(c, "This logic code already creted outside of the current transaction")
					return nil
				}
			}, nil)
			if err != nil {
				log.Errorf(c, fmt.Errorf("%w: transaction failed", err).Error())
			} else if created {
				return int32(code), nil
			}
		} else if err != nil {
			log.Errorf(c, fmt.Errorf("failed to get entity: %w", err).Error())
		}
	}
	return 0, errors.New("Failed to create new loginc code")
}

func (LoginCodeDalGae) ClaimLoginCode(c context.Context, code int32) (userID int64, err error) {
	err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
		key := NewLoginCodeKey(c, code)
		var entity models.LoginCodeEntity
		if err := gaedb.Get(c, key, &entity); err != nil {
			if err == datastore.ErrNoSuchEntity {
				return err
			} else {
				return fmt.Errorf("failed to get %v: %w", key, err)
			}
		}
		if entity.Created.Add(time.Minute).Before(time.Now()) {
			return models.ErrLoginCodeExpired
		}
		var emptyTime time.Time
		if entity.Claimed == emptyTime {
			return models.ErrLoginCodeAlreadyClaimed
		}
		entity.Claimed = time.Now()
		if _, err := gaedb.Put(c, key, &entity); err != nil {
			return fmt.Errorf("failed to save %v: %w", key, err)
		}
		userID = entity.UserID
		return nil
	}, nil)
	return
}
