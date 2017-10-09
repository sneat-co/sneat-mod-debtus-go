package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"strings"
	"sync"
	"github.com/strongo/app/gaedb"
)

type TgChatDalGae struct {
}

func NewTgChatDalGae() TgChatDalGae {
	return TgChatDalGae{}
}

func (_ TgChatDalGae) DoSomething(c context.Context, userTask *sync.WaitGroup, currency string, tgChatID int64, authInfo auth.AuthInfo, user models.AppUser, sendToTelegram func(tgChat telegram_bot.TelegramChatEntityBase) error) (err error) {
	var isSentToTelegram bool // Needed in case of failed to save to DB and is auto-retry
	tgChatKey := gaedb.NewKey(c, telegram_bot.TelegramChatKind, "", tgChatID, nil)
	if err = gaedb.RunInTransaction(c, func(tc context.Context) (err error) {
		var tgChat telegram_bot.TelegramChatEntityBase

		if err = gaedb.Get(tc, tgChatKey, &tgChat); err != nil {
			return errors.Wrapf(err, "Failed to get Telegram chat entity by key=%v", tgChatKey)
		}
		if tgChat.BotID == "" {
			log.Errorf(c, "Data inconsitence issue - TgChat(%v).BotID is empty string", tgChatID)
			if strings.Contains(authInfo.Issuer, ":") {
				issuer := strings.Split(authInfo.Issuer, ":")
				if strings.ToLower(issuer[0]) == "telegram" {
					tgChat.BotID = issuer[1]
					log.Infof(c, "Data inconsitence fixed, set to: %v", tgChat.BotID)
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
