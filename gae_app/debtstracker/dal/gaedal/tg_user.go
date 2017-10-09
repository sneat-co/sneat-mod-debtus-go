package gaedal

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type TgUserDalGae struct {
}

func NewTgUserDalGae() TgUserDalGae {
	return TgUserDalGae{}
}

func (_ TgUserDalGae) FindByUserName(c context.Context, userName string) (tgUsers []telegram_bot.TelegramUser, err error) {
	var tgUserEntities []telegram_bot.TelegramUserEntity

	query := datastore.NewQuery(telegram_bot.TelegramUserKind)
	query = query.Filter("UserName =", userName)

	var keys []*datastore.Key
	keys, err = query.GetAll(c, &tgUserEntities)

	if err != nil {
		return
	}

	tgUsers = make([]telegram_bot.TelegramUser, len(keys))
	for i, entity := range tgUserEntities {
		tgUsers[i] = telegram_bot.TelegramUser{ID: keys[i].IntID(), TelegramUserEntity: entity}
	}
	return
}
