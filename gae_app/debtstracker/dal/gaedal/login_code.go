package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"math/rand"
	"time"
)

type LoginCodeDalGae struct {
}

func NewLoginCodeDalGae() LoginCodeDalGae {
	return LoginCodeDalGae{}
}

func NewLoginCodeKey(c context.Context, code int32) *datastore.Key {
	return gaedb.NewKey(c, models.LoginCodeKind, "", int64(code), nil)
}

func (_ LoginCodeDalGae) NewLoginCode(c context.Context, userID int64) (int32, error) {
	var code int32
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 1; i < 20; i++ {
		code = random.Int31n(99999) + 1
		key := NewLoginCodeKey(c, code)
		if err := gaedb.Get(c, key, &models.LoginCodeEntity{}); err == datastore.ErrNoSuchEntity {
			var created bool
			err = dal.DB.RunInTransaction(c, func(c context.Context) error {
				var entity models.LoginCodeEntity
				if err := gaedb.Get(c, key, &entity); err == datastore.ErrNoSuchEntity || entity.Created.Add(time.Hour).Before(time.Now()) {
					entity = models.LoginCodeEntity{
						Created: time.Now(),
						UserID:  userID,
					}
					if _, err = gaedb.Put(c, key, &entity); err != nil {
						log.Errorf(c, errors.Wrapf(err, "Failed to save %v", key).Error())
					}
					created = true
					return nil
				} else if err != nil {
					return errors.Wrap(err, "Failed to get entity withing transaction")
				} else {
					log.Warningf(c, "This logic code already creted outside of the current transaction")
					return nil
				}
			}, nil)
			if err != nil {
				log.Errorf(c, errors.Wrap(err, "Transaction failed").Error())
			} else if created {
				return int32(code), nil
			}
		} else if err != nil {
			log.Errorf(c, errors.Wrap(err, "Failed to get entity").Error())
		}
	}
	return 0, errors.New("Failed to create new loginc code")
}

func (_ LoginCodeDalGae) ClaimLoginCode(c context.Context, code int32) (userID int64, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		key := NewLoginCodeKey(c, code)
		var entity models.LoginCodeEntity
		if err := gaedb.Get(c, key, &entity); err != nil {
			if err == datastore.ErrNoSuchEntity {
				return err
			} else {
				return errors.Wrapf(err, "Failed to get %v", key)
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
			return errors.Wrapf(err, "Failed to save %v", key)
		}
		userID = entity.UserID
		return nil
	}, nil)
	return
}
