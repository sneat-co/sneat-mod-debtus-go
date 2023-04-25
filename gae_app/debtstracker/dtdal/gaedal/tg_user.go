package gaedal

import (
	"context"
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"reflect"
)

type TgUserDalGae struct {
}

func NewTgUserDalGae() TgUserDalGae {
	return TgUserDalGae{}
}

func (TgUserDalGae) FindByUserName(c context.Context, tx dal.ReadSession, userName string) (tgUsers []tgstore.TgUser, err error) {
	if tx == nil {
		tx, err = facade.GetDatabase(c)
		if err != nil {
			return
		}
	}
	q := dal.From(tgstore.BotUserCollection).
		WhereField("UserName", dal.Equal, userName)

	query := q.SelectInto(func() dal.Record {
		return dal.NewRecordWithIncompleteKey(tgstore.BotUserCollection, reflect.Int, new(tgstore.TgBotUserData))
	})
	var records []dal.Record

	if records, err = tx.QueryAllRecords(c, query); err != nil {
		return
	}
	tgUsers = make([]tgstore.TgUser, len(records))
	for i, r := range records {
		tgUsers[i] = tgstore.TgUser{
			WithID: record.NewWithID(r.Key().ID.(int64), r.Key(), r.Data),
			Data:   r.Data().(*tgstore.TgBotUserData),
		}
	}
	return
}
