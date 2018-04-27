package gaedal

import (
	"strings"
	"sync"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
)

type TgChatDalGae struct {
}

func NewTgChatDalGae() TgChatDalGae {
	return TgChatDalGae{}
}

func (TgChatDalGae) GetTgChatByID(c context.Context, tgBotID string, tgChatID int64) (tgChat models.TelegramChat, err error) {
	tgChat.SetID(tgBotID, tgChatID)
	err = dal.DB.Get(c, &tgChat)
	return
}

func (TgChatDalGae) SaveTgChat(c context.Context, tgChat models.TelegramChat) error {
	return dal.DB.Update(c, &tgChat)
}

func (TgChatDalGae) /* TODO: rename properly! */ DoSomething(c context.Context,
	userTask *sync.WaitGroup, currency string, tgChatID int64, authInfo auth.AuthInfo, user models.AppUser,
	sendToTelegram func(tgChat telegram.TgChatEntityBase) error,
) (err error) {
	var isSentToTelegram bool // Needed in case of failed to save to DB and is auto-retry
	tgChatKey := gaedb.NewKey(c, telegram.ChatKind, "", tgChatID, nil)
	if err = gaedb.RunInTransaction(c, func(tc context.Context) (err error) {
		var tgChat telegram.TgChatEntityBase

		if err = gaedb.Get(tc, tgChatKey, &tgChat); err != nil {
			return errors.Wrapf(err, "Failed to get Telegram chat entity by key=%v", tgChatKey)
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
			return errors.Wrap(err, "Failed to save Telegram chat entity to datastore")
		}
		return err
	}, nil); err != nil {
		err = errors.Wrap(err, "TgChatDalGae.DoSomething() transaction failed")
	}
	return
}
