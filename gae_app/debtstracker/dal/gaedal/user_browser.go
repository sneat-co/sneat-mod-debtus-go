package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"strings"
	"time"
	"github.com/strongo/app/gaedb"
)

func newUserBrowserIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.UserBrowserKind, nil)
}

type UserBrowserDalGae struct {
}

func NewUserBrowserDalGae() UserBrowserDalGae {
	return UserBrowserDalGae{}
}

func (_ UserBrowserDalGae) insertUserBrowser(c context.Context, entity *models.UserBrowserEntity) (userBrowser models.UserBrowser, err error) {
	var key *datastore.Key
	if key, err = gaedb.Put(c, newUserBrowserIncompleteKey(c), &entity); err != nil {
		return
	}
	userBrowser = models.UserBrowser{ID: key.IntID(), UserBrowserEntity: entity}
	return
}

func (userBrowserDalGae UserBrowserDalGae) SaveUserBrowser(c context.Context, userID int64, userAgent string) (userBrowser models.UserBrowser, err error) {
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		panic("Missign required parameter userAgent")
	}
	const limit = 1
	query := datastore.NewQuery(models.UserBrowserKind).Limit(limit)
	query = query.Filter("AppUserIntID =", userID).Filter("UserAgent =", userAgent)
	userBrowsers := make([]models.UserBrowserEntity, 0, limit)
	var keys []*datastore.Key
	if keys, err = query.GetAll(c, &userBrowsers); err != nil {
		err = errors.New(fmt.Sprintf("Failed to query UserBrowser: %v", err))
		return
	} else {
		switch len(keys) {
		case 0:
			ub := models.UserBrowserEntity{
				UserID:      userID,
				UserAgent:   userAgent,
				LastUpdated: time.Now(),
			}
			userBrowser, err = userBrowserDalGae.insertUserBrowser(c, &ub)
			return
		case 1:
			userBrowser := userBrowsers[0]
			if userBrowser.LastUpdated.Before(time.Now().Add(-24 * time.Hour)) {
				gaedb.RunInTransaction(c, func(c context.Context) error {
					key := keys[0]
					if err := gaedb.Get(c, key, &userBrowser); err != nil {
						return err
					}
					if userBrowser.LastUpdated.Before(time.Now().Add(-time.Hour)) {
						userBrowser.LastUpdated = time.Now()
						_, err = gaedb.Put(c, key, &userBrowser)
					}
					return err
				}, nil)
			}
		default:
			log.Errorf(c, "Loaded too many entities: %v", len(keys))
		}
		return
	}
}
