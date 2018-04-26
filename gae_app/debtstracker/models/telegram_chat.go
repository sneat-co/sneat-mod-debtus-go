package models

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

type TelegramChat struct {
	telegram.TgChatBase
	*DtTelegramChatEntity
}

var _ db.EntityHolder = (*TelegramChat)(nil)

func (TelegramChat) Kind() string {
	return telegram.ChatKind
}

func (tgChat TelegramChat) Entity() interface{} {
	return tgChat.DtTelegramChatEntity
}

func (TelegramChat) NewEntity() interface{} {
	return new(DtTelegramChatEntity)
}

func (tgChat *TelegramChat) SetEntity(entity interface{}) {
	if entity == nil {
		tgChat.DtTelegramChatEntity = nil
	} else {
		tgChat.DtTelegramChatEntity = entity.(*DtTelegramChatEntity)
	}
}

type DtTelegramChatEntity struct {
	UserGroupID string `datastore:",index"` // Do index
	telegram.TgChatEntityBase
}

func (entity *DtTelegramChatEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *DtTelegramChatEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return properties, err
	}
	if properties, err = entity.TgChatEntityBase.CleanProperties(properties); err != nil {
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
