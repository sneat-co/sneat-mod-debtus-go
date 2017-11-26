package facade

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"sort"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/memcache"
)

type refererFacade struct {
}

var Referer = refererFacade{}

const lastTgReferrers = "lastTgReferrers"

var errAlreadyReferred = errors.New("already referred")

func setUserReferrer(c context.Context, userID int64, referredBy string) (err error) {
	userChanged := false
	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		user, err := dal.User.GetUserByID(c, userID)
		if err != nil {
			return err
		}
		if user.ReferredBy != "" {
			log.Debugf(c, "already referred")
			return nil
		}
		user.ReferredBy = referredBy
		userChanged = true
		return dal.User.SaveUser(c, user)
	}, db.CrossGroupTransaction); err != nil {
		log.Errorf(c, "failed to check & update user: %v", err)
		return err
	}
	if userChanged {
		log.Infof(c, "User's referrer saved")
	}
	return nil
}

var delayedSetUserReferrer = delay.Func("setUserReferrer", setUserReferrer)

func delaySetUserReferrer(c context.Context, userID int64, referredBy string) (err error) {
	return gae.CallDelayFuncWithDelay(c, time.Second/2, common.QUEUE_USERS, "set-user-referrer", delayedSetUserReferrer, userID, referredBy)
}

func (f refererFacade) AddTelegramReferrer(c context.Context, userID int64, tgUsername, botID string) {
	referer := models.Referer{
		RefererEntity: &models.RefererEntity{
			Platform:   "tg",
			ReferredTo: botID,
			DtCreated:  time.Now(),
			ReferredBy: tgUsername,
		},
	}
	go func() {
		user, err := dal.User.GetUserByID(c, userID)
		if err != nil {
			log.Errorf(c, err.Error())
			return
		}
		if user.ReferredBy != "" {
			log.Debugf(c, "already referred")
			return
		}
		delaySetUserReferrer(c, userID, "tg:"+tgUsername)
		item, err := memcache.Get(c, lastTgReferrers)
		var isLocked bool
		if err == memcache.ErrCacheMiss {
			item = f.lockMemcacheItem(c)
			isLocked = true
		}
		if err := dal.DB.InsertWithRandomIntID(c, &referer); err != nil {
			log.Errorf(c, "failed to insert referer to DB: %v", err)
		}
		if err != nil {
			log.Warningf(c, "failed to get last-tg-referrers from memcache")
		}
		if item == nil {
			if err = memcache.Delete(c, lastTgReferrers); err != nil {
				log.Warningf(c, "Failed to clear memcache item: %v", err) // TODO: add a queue task to remove?
				return
			}
		} else {
			var tgUsernames []string
			if isLocked {
				tgUsernames = []string{tgUsername}
			} else {
				tgUsernames = append(strings.Split(string(item.Value), ","), tgUsername)
				if len(tgUsernames) > 100 {
					tgUsernames = tgUsernames[:100]
				}
			}
			item.Value = []byte(strings.Join(tgUsernames, ","))
			if err = memcache.CompareAndSwap(c, item); err != nil {
				if err = memcache.Delete(c, lastTgReferrers); err != nil {
					log.Warningf(c, "failed to delete '%v' from memcache", lastTgReferrers)
				}
			}
		}
	}()
}

func (refererFacade) lockMemcacheItem(c context.Context) (item *memcache.Item) {
	lock := make([]byte, 9)
	lock[0] = []byte("_")[0]
	binary.LittleEndian.PutUint64(lock[1:], rand.Uint64())
	item = &memcache.Item{
		Key:        lastTgReferrers,
		Value:      lock,
		Expiration: time.Second * 10,
	}

	if err := memcache.Set(c, item); err == nil {
		if item, err = memcache.Get(c, item.Key); err != nil {
			log.Warningf(c, "memcache error: %v", err)
			item = nil
		} else if !bytes.Equal(lock, item.Value) {
			item = nil
		}
	}
	return
}

func (f refererFacade) TopTelegramReferrers(c context.Context, botID string, limit int) (topTelegramReferrers []string, err error) {
	var item *memcache.Item
	var tgUsernames []string

	if item, err = memcache.Get(c, lastTgReferrers); err == nil && item != nil && len(item.Value) > 0 && item.Value[0] != []byte("_")[0] {
		tgUsernames = strings.Split(string(item.Value), ",")
		item = nil
	} else {
		item = f.lockMemcacheItem(c)
		query := datastore.NewQuery(models.RefererKind).Filter("p =", "tg").Filter("to =", botID).Order("-t").Limit(100)
		iterator := query.Run(c)
		refererEntity := new(models.RefererEntity)
		for {
			if _, err = iterator.Next(refererEntity); err != nil {
				if err == datastore.Done {
					err = nil
					break
				}
				return
			}
			tgUsernames = append(tgUsernames, refererEntity.ReferredBy)
		}
	}
	counts := make(map[string]int, len(tgUsernames))
	for _, tgUsername := range tgUsernames {
		counts[tgUsername] += 1
	}

	count := len(counts)
	if count > limit {
		count = limit
	}

	topTelegramReferrers = rankByCount(counts, limit)
	if item != nil {
		item.Value = []byte(strings.Join(tgUsernames, ","))
		item.Expiration = 0
		memcache.CompareAndSwap(c, item)
	} else {
		v := []byte(strings.Join(tgUsernames, ","))
		item = &memcache.Item{
			Key:   lastTgReferrers,
			Value: v,
		}
		if err = memcache.Set(c, item); err != nil {
			if err = memcache.Delete(c, item.Key); err != nil {
				log.Warningf(c, "Failed to clear memcache: %v", err)
				err = nil
			}
		}
	}

	return
}

func rankByCount(countsByName map[string]int, limit int) (names []string) {
	pl := make(PairList, len(countsByName))
	i := 0
	for k, v := range countsByName {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	if len(pl) <= limit {
		names = make([]string, len(pl))
	} else {
		names = make([]string, limit)
	}
	for i := range pl {
		names[i] = pl[i].Key
	}
	return
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
