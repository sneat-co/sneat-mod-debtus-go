package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"github.com/strongo/db"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/log"
	users "github.com/strongo/app/user"
	"strconv"
	"strings"
	"fmt"
)

type verifyTelegramUserAccounts struct {
	asyncMapper
	entity *models.DtTelegramChatEntity
}

func (m *verifyTelegramUserAccounts) Make() interface{} {
	m.entity = new(models.DtTelegramChatEntity)
	return m.entity
}

func (m *verifyTelegramUserAccounts) Query(r *http.Request) (query *mapper.Query, err error) {
	var filtered bool
	if query, filtered, err = filterByStrID(r, telegram_bot.TelegramChatKind, "tgchat"); err != nil {
		return
	} else {
		paramsCount := len(r.URL.Query())
		if filtered {
			paramsCount -= 1
		}
		if paramsCount != 1 {
			err = errors.New("unexpected params: " + r.URL.RawQuery)
		}
	}
	return
}

func (m *verifyTelegramUserAccounts) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	entity := *m.entity
	if key.StringID() == "" {
		if key.IntID() != 0 {
			counters.Increment("integer-keys", 1)
			return m.startWorker(c, counters, func() Worker {
				return func(counters *asyncCounters) error {
					return m.dealWithIntKey(c, counters, key, &entity)
				}
			})
			return
		}
	} else {
		tgChat := models.TelegramChat{DtTelegramChatEntity: &entity}
		tgChat.StringID = db.NewStrID(key.StringID())
		return m.startWorker(c, counters, func() Worker {
			return func(counters *asyncCounters) error {
				return m.processTelegramChat(c, tgChat, counters)
			}
		})
	}
	return
}

func (m *verifyTelegramUserAccounts) dealWithIntKey(c context.Context, counters *asyncCounters, key *datastore.Key, tgChatEntity *models.DtTelegramChatEntity) (err error) {
	if tgChatEntity.BotID == "" {
		counters.Increment("empty_BotID_count", 1)
		if err = datastore.Delete(c, key); err != nil {
			log.Errorf(c, "failed to delete %v: %v", key.IntID(), err)
			return nil
		}
		counters.Increment("empty_BotID_deleted", 1)
	}
	var tgChat models.TelegramChat
	if tgChat, err = dal.TgChat.GetTgChatByID(c, tgChatEntity.BotID, tgChatEntity.TelegramUserID); err != nil {
		if db.IsNotFound(err) {
			tgChat.SetID(tgChatEntity.BotID, tgChatEntity.TelegramUserID)
			tgChat.SetEntity(tgChatEntity)
			if err = dal.DB.Update(c, &tgChat); err != nil {
				log.Errorf(c, "failed to created entity with fixed key %v: %v", tgChat.ID, err)
				return nil
			}
			if err = datastore.Delete(c, key); err != nil {
				log.Errorf(c, "failed to delete migrated %v: %v", key.IntID(), err)
				return nil
			}
			counters.Increment("migrated", 1)
		}
	} else if tgChat.BotID == tgChatEntity.BotID && tgChat.TelegramUserID == tgChatEntity.TelegramUserID {
		if err = datastore.Delete(c, key); err != nil {
			log.Errorf(c, "failed to delete already migrated %v: %v", key.IntID(), err)
			return nil
		}
		counters.Increment("already_migrated_so_deleted", 1)
	} else {
		counters.Increment("mismatches", 1)
		if tgChat.BotID != tgChatEntity.BotID {
			log.Warningf(c, "%v: tgChat.BotID != tgChatEntity.BotID: %v != %v", key.IntID(), tgChat.BotID, tgChatEntity.BotID)
		} else if tgChat.TelegramUserID != tgChatEntity.TelegramUserID {
			log.Warningf(c, "%v: tgChat.TelegramUserID != tgChatEntity.TelegramUserID: %v != %v", key.IntID(), tgChat.TelegramUserID, tgChatEntity.TelegramUserID)
		}
	}
	return
}

func (m *verifyTelegramUserAccounts) processTelegramChat(c context.Context, tgChat models.TelegramChat, counters *asyncCounters) (err error) {
	var (
		user        models.AppUser
		userChanged bool
	)
	if tgChat.BotID == "" || tgChat.TelegramUserID == 0 {
		log.Warningf(c, "TgChat(%v) => BotID=%v, TelegramUserID=%v", tgChat.ID, tgChat.TelegramUserID)
		if strings.Contains(tgChat.ID, ":") {
			botID := strings.Split(tgChat.ID, ":")[0]
			tgUserID := tgChat.TelegramUserID
			if tgUserID == 0 {
				if tgUserID, err = strconv.ParseInt(strings.Split(tgChat.ID, ":")[1], 10, 64); err != nil {
					return
				}
			}
			if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
				if tgChat, err = dal.TgChat.GetTgChatByID(c, botID, tgUserID); err != nil {
					return
				}
				tgChat.TelegramUserID = tgUserID
				tgChat.BotID = botID
				return dal.DB.Update(c, &tgChat)
			}, db.CrossGroupTransaction); err != nil {
				log.Errorf(c, "Failed to fix TgChat(%v): %v", tgChat.ID, err)
				err = nil
				return
			}
			log.Infof(c, "Fixed TgChat(%v)", tgChat.ID)
		} else {
			return
		}
	}
	if tgChat.AppUserIntID == 0 {
		log.Warningf(c, "TgChat(%v).AppUserIntID == 0", tgChat.ID)
		return
	}
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if user, err = dal.User.GetUserByID(c, tgChat.AppUserIntID); err != nil {
			if db.IsNotFound(err) {
				log.Errorf(c, "Failed to process %v: %v", tgChat.ID, err)
				err = nil
			}
			return
		}
		telegramAccounts := user.GetTelegramAccounts()
		tgChatStrID := strconv.FormatInt(tgChat.TelegramUserID, 10)
		for _, ua := range telegramAccounts {
			if ua.ID == tgChatStrID {
				if ua.App == tgChat.BotID {
					goto userAccountFound
				} else if ua.App == "" {
					//log.Debugf(c, "will be fixed")
					user.RemoveAccount(ua)
					ua.App = tgChat.BotID
					userChanged = user.AddAccount(ua) || userChanged
					goto userAccountFound
				}
			}
		}
		userChanged = user.AddAccount(users.Account{
			ID:       strconv.FormatInt(tgChat.TelegramUserID, 10),
			App:      tgChat.BotID,
			Provider: telegram_bot.TelegramPlatformID,
		}) || userChanged
	userAccountFound:
		if userChanged {
			//log.Debugf(c, "user changed %v", user.ID)
			defer func() {
				if r := recover(); r != nil {
					log.Errorf(c, "panic on saving user %v: %v", user.ID, r)
					err = fmt.Errorf("panic on saving user %v: %v", user.ID, r)
				}
			}()
			if err = dal.User.SaveUser(c, user); err != nil {
				return
			}
		//} else {
		//	log.Debugf(c, "user NOT changed %v", user.ID)
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		counters.Increment("failed", 1)
		return
	} else if userChanged {
		log.Infof(c, "User %v fixed", user.ID)
		counters.Increment("users-changed", 1)
	}
	return
}
