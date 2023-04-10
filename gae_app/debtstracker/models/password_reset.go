package models

import (
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app/user"
)

const PasswordResetKind = "PwdRst"

type PasswordReset struct {
	record.WithID[int]
	*PasswordResetEntity
}

//var _ db.EntityHolder = (*PasswordReset)(nil)

type PasswordResetEntity struct {
	Email  string
	Status string
	user.OwnedByUserWithIntID
}

func (PasswordReset) Kind() string {
	return PasswordResetKind
}

//func (record PasswordReset) IntID() int64 {
//	return record.ID
//}

func (record PasswordReset) Entity() interface{} {
	return record.PasswordResetEntity
}

func (PasswordReset) NewEntity() interface{} {
	return new(PasswordResetEntity)
}

func (record *PasswordReset) SetEntity(entity interface{}) {
	if entity == nil {
		record.PasswordResetEntity = nil
	} else {
		record.PasswordResetEntity = entity.(*PasswordResetEntity)
	}
}

//func (entity *PasswordResetEntity) Save() (properties []datastore.Property, err error) {
//	if properties, err = datastore.SaveStruct(entity); err != nil {
//		return
//	}
//	return gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
//		"DtUpdated": gaedb.IsZeroTime,
//		"Email":     gaedb.IsEmptyString,
//	})
//}
