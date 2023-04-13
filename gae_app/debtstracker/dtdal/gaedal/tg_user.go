package gaedal

import (
	"context"
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"google.golang.org/appengine/v2/datastore"
)

type TgUserDalGae struct {
}

func NewTgUserDalGae() TgUserDalGae {
	return TgUserDalGae{}
}

func (TgUserDalGae) FindByUserName(c context.Context, userName string) (tgUsers []tgstore.TgUser, err error) {
	var tgUserDatas []tgstore.TgBotUserData

	query := datastore.NewQuery(tgstore.BotUserCollection)
	query = query.Filter("UserName =", userName)

	var keys []*datastore.Key
	keys, err = query.GetAll(c, &tgUserDatas)

	if err != nil {
		return
	}

	tgUsers = make([]tgstore.TgUser, len(keys))
	for i, entity := range tgUserDatas {
		tgUsers[i] = tgstore.TgUser{ID: keys[i].IntID(), TgUserEntity: entity}
	}
	return
}
