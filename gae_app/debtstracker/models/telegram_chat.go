package models

import (
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
)

type TelegramChat struct {
	tgstore.Chat
	//tgstore.ChatEntity
	Data *DebtusTelegramChatData
}

//var _ db.EntityHolder = (*TelegramChat)(nil)

//func (TelegramChat) Kind() string {
//	return telegram.ChatKind
//}

//func (tgChat TelegramChat) Entity() interface{} {
//	return tgChat.DebtusTelegramChatData
//}

//func (TelegramChat) NewEntity() interface{} {
//	return new(DebtusTelegramChatData)
//}

//func (tgChat *TelegramChat) SetEntity(entity interface{}) {
//	if entity == nil {
//		tgChat.DebtusTelegramChatData = nil
//	} else {
//		tgChat.DebtusTelegramChatData = entity.(*DebtusTelegramChatData)
//	}
//}

// DebtusTelegramChatData is a data structure for storing debtus data related to specific telegram chat
type DebtusTelegramChatData struct {
	tgstore.TgChatBase
	DebtusChatData
}

func (entity *DebtusTelegramChatData) Validate() (err error) {
	//if properties, err = datastore.SaveStruct(entity); err != nil {
	//	return properties, err
	//}
	//if properties, err = entity.TgChatEntityBase.CleanProperties(properties); err != nil {
	//	return
	//}
	//if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
	//	"GetUserGroupID":   gaedb.IsEmptyString,
	//	"TgChatInstanceID": gaedb.IsEmptyString,
	//}); err != nil {
	//	return
	//}
	return
}
