package gaedal

import (
	"fmt"
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"strconv"
	"strings"
	"sync"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
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
		TgChat: tgstore.TgChat{
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
	sendToTelegram func(tgChat tgstore.TgChat) error,
) (err error) {
	var isSentToTelegram bool // Needed in case of failed to save to DB and is auto-retry
	debtusTgChatData := &models.DebtusTelegramChatData{}

	debtusTgChat := models.DebtusTelegramChat{
		TgChat: tgstore.NewTgChat(strconv.FormatInt(tgChatID, 10), debtusTgChatData),
	}

	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}

	if err = db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) (err error) {
		if err = tx.Get(tc, debtusTgChat.Record); err != nil {
			return err
		}
		if debtusTgChat.Data.BotID == "" {
			log.Errorf(c, "Data inconsistency issue - TgChat(%v).BotID is empty string", tgChatID)
			if strings.Contains(authInfo.Issuer, ":") {
				issuer := strings.Split(authInfo.Issuer, ":")
				if strings.ToLower(issuer[0]) == "telegram" {
					debtusTgChat.Data.BotID = issuer[1]
					log.Infof(c, "Data inconsistency fixed, set to: %v", debtusTgChat.Data.BotID)
				}
			}
		}
		debtusTgChat.Data.AddWizardParam("currency", string(currency))

		if !isSentToTelegram {
			if err = sendToTelegram(debtusTgChat.TgChat); err != nil { // This is some serious architecture sheet. Too sleepy to make it right, just make it working.
				return err
			}
			isSentToTelegram = true
		}
		if err = tx.Set(tc, debtusTgChat.Record); err != nil {
			return fmt.Errorf("failed to save Telegram chat record to db: %w", err)
		}
		return err
	}, nil); err != nil {
		err = fmt.Errorf("method TgChatDalGae.DoSomething() transaction failed: %w", err)
	}
	return
}
