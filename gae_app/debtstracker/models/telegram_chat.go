package models

import (
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/dal-go/dalgo/dal"
	"reflect"
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

func NewDebtusTelegramChatRecord() dal.Record {
	return dal.NewRecordWithIncompleteKey(tgstore.TgChatCollection, reflect.String, new(DebtusTelegramChatData))
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
