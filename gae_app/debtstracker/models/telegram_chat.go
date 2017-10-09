package models

import (
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/datastore"
)

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
		"UserGroupID":      gaedb.IsZeroInt,
		"TgChatInstanceID": gaedb.IsEmptyString,
	}); err != nil {
		return
	}
	return
}
