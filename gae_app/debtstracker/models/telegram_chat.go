package models

import (
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
)

type DebtusTelegramChat struct {
	tgstore.TgChat
	//tgstore.ChatEntity
	Data *DebtusTelegramChatData
}

var _ tgstore.TgChatData = (*DebtusTelegramChatData)(nil)

// DebtusTelegramChatData is a data structure for storing debtus data related to specific telegram chat
type DebtusTelegramChatData struct {
	tgstore.TgChatBase
	DebtusChatData
}

func (v *DebtusTelegramChatData) BaseChatData() *tgstore.TgChatBase {
	return &v.TgChatBase
}

func (v *DebtusTelegramChatData) Validate() (err error) {
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
