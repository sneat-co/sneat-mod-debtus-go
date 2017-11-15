package models

import (
	"github.com/strongo/app/gaedb"
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/db"
)

type TelegramChat struct {
	db.IntegerID
	*DtTelegramChatEntity
}

var _ db.EntityHolder = (*TelegramChat)(nil)

func (TelegramChat) Kind() string {
	return telegram_bot.TelegramChatKind
}

func (tgChat TelegramChat) Entity() interface{} {
	return tgChat.DtTelegramChatEntity
}

func (tgChat *TelegramChat) SetEntity(entity interface{}) {
	tgChat.DtTelegramChatEntity = entity.(*DtTelegramChatEntity)
}

type DtTelegramChatEntity struct {
	UserGroupID string `datastore:",index"` // Do index
	telegram_bot.TelegramChatEntityBase
}

func (entity *DtTelegramChatEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *DtTelegramChatEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return properties, err
	}
	if properties, err = entity.TelegramChatEntityBase.CleanProperties(properties); err != nil {
		return
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"UserGroupID":      gaedb.IsEmptyString,
		"TgChatInstanceID": gaedb.IsEmptyString,
	}); err != nil {
		return
	}
	return
}
