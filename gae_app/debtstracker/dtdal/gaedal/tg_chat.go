package gaedal

import (
	"fmt"
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"strings"
	"sync"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
)

type TgChatDalGae struct {
}

func NewTgChatDalGae() TgChatDalGae {
	return TgChatDalGae{}
}

func (TgChatDalGae) GetTgChatByID(c context.Context, tgBotID string, tgChatID int64) (tgChat models.DebtusTelegramChat, err error) {
	tgChatFullID := fmt.Sprintf("%s:%d", tgBotID, tgChatID)
	key := dal.NewKeyWithID(tgstore.TgChatCollection, tgChatFullID)
	data := new(models.DebtusTelegramChatData)
	tgChat = models.DebtusTelegramChat{
		Chat: tgstore.Chat{
			WithID: record.NewWithID[string](tgChatFullID, key, data),
		},
		Data: data,
	}
	//tgChat.SetID(tgBotID, tgChatID)

	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}
	err = db.Get(c, tgChat.Record)
	return
}

//func (TgChatDalGae) SaveTgChat(c context.Context, tgChat models.DebtusTelegramChat) error {
//	return dtdal.DB.Update(c, &tgChat)
//}

func (TgChatDalGae) /* TODO: rename properly! */ DoSomething(c context.Context,
	userTask *sync.WaitGroup, currency string, tgChatID int64, authInfo auth.AuthInfo, user models.AppUser,
	sendToTelegram func(tgChatBase tgstore.TgChatBase) error,
) (err error) {
	var isSentToTelegram bool // Needed in case of failed to save to DB and is auto-retry
	tgChatKey := gaedb.NewKey(c, tgstore.TgChatCollection, "", tgChatID, nil)
	if err = gaedb.RunInTransaction(c, func(tc context.Context) (err error) {
		var tgChat tgstore.TgChatBase

		if err = gaedb.Get(tc, tgChatKey, &tgChat); err != nil {
			return fmt.Errorf("failed to get Telegram chat entity by key=%v: %w", tgChatKey, err)
		}
		if tgChat.BotID == "" {
			log.Errorf(c, "Data inconsistency issue - TgChat(%v).BotID is empty string", tgChatID)
			if strings.Contains(authInfo.Issuer, ":") {
				issuer := strings.Split(authInfo.Issuer, ":")
				if strings.ToLower(issuer[0]) == "telegram" {
					tgChat.BotID = issuer[1]
					log.Infof(c, "Data inconsistency fixed, set to: %v", tgChat.BotID)
				}
			}
		}
		tgChat.AddWizardParam("currency", string(currency))

		if !isSentToTelegram {
			if err = sendToTelegram(tgChat); err != nil { // This is some serious architecture sheet. Too sleepy to make it right, just make it working.
				return err
			}
			isSentToTelegram = true
		}
		if _, err = gaedb.Put(tc, tgChatKey, &tgChat); err != nil {
			return fmt.Errorf("failed to save Telegram chat entity to datastore: %w", err)
		}
		return err
	}, nil); err != nil {
		err = fmt.Errorf("method TgChatDalGae.DoSomething() transaction failed: %w", err)
	}
	return
}
