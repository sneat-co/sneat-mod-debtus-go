package models

import "github.com/strongo/app/db"

const TgGroupKind = "TgGroup"

type TgGroup struct {
	db.IntegerID
	*TgGroupEntity
}

var _ db.EntityHolder = (*TgGroup)(nil)

type TgGroupEntity struct {
	UserGroupID string `datastore:",noindex"`
}

func (TgGroup) Kind() string {
	return TgGroupKind
}

func (tgGroup TgGroup) Entity() interface{} {
	return tgGroup.TgGroupEntity
}

func (tgGroup *TgGroup) SetEntity(entity interface{}) {
	tgGroup.TgGroupEntity, _ = entity.(*TgGroupEntity)
}
