package models

import (
	"github.com/dal-go/dalgo/record"
)

const TgGroupKind = "TgGroup"

type TgGroup struct {
	record.WithID[string]
	*TgGroupEntity
}

//var _ db.EntityHolder = (*TgGroup)(nil)

type TgGroupEntity struct {
	UserGroupID string `datastore:",noindex"`
}

func (TgGroup) Kind() string {
	return TgGroupKind
}

func (tgGroup TgGroup) Entity() interface{} {
	return tgGroup.TgGroupEntity
}

func (tgGroup TgGroup) NewEntity() interface{} {
	return new(TgGroupEntity)
}

func (tgGroup *TgGroup) SetEntity(entity interface{}) {
	if entity == nil {
		tgGroup.TgGroupEntity = nil
	} else {
		tgGroup.TgGroupEntity = entity.(*TgGroupEntity)
	}
}
