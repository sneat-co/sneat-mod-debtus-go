package gaedal

import (
	"context"
	"github.com/bots-go-framework/bots-fw-telegram"
	"google.golang.org/appengine/v2/datastore"
)

type TgUserDalGae struct {
}

func NewTgUserDalGae() TgUserDalGae {
	return TgUserDalGae{}
}

func (TgUserDalGae) FindByUserName(c context.Context, userName string) (tgUsers []telegram.TgUser, err error) {
	var tgUserEntities []telegram.TgUserEntity

	query := datastore.NewQuery(telegram.TgUserKind)
	query = query.Filter("UserName =", userName)

	var keys []*datastore.Key
	keys, err = query.GetAll(c, &tgUserEntities)

	if err != nil {
		return
	}

	tgUsers = make([]telegram.TgUser, len(keys))
	for i, entity := range tgUserEntities {
		tgUsers[i] = telegram.TgUser{ID: keys[i].IntID(), TgUserEntity: entity}
	}
	return
}
