package models

import (
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
)

type DebtusTelegramChat struct {
	tgstore.Chat
	//tgstore.ChatEntity
	Data *DebtusTelegramChatData
}

//var _ db.EntityHolder = (*DebtusTelegramChat)(nil)

//func (DebtusTelegramChat) Kind() string {
//	return telegram.ChatKind
//}

//func (tgChat DebtusTelegramChat) Entity() interface{} {
//	return tgChat.DebtusTelegramChatData
//}

//func (DebtusTelegramChat) NewEntity() interface{} {
//	return new(DebtusTelegramChatData)
//}

//func (tgChat *DebtusTelegramChat) SetEntity(entity interface{}) {
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
