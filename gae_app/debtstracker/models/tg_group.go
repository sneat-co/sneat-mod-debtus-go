package models

import "github.com/strongo/app/db"

const TgGroupKind = "TgGroup"

type TgGroup struct {
	db.NoStrID
	ID int64
	*TgGroupEntity
}

var _ db.EntityHolder = (*TgGroup)(nil)

type TgGroupEntity struct {
	UserGroupID string `datastore:",noindex"`
}

func (TgGroup) Kind() string {
	return TgGroupKind
}

func (tgGroup TgGroup) IntID() int64 {
	return tgGroup.ID
}

func (tgGroup TgGroup) Entity() interface{} {
	return tgGroup.TgGroupEntity
}

func (tgGroup *TgGroup) SetEntity(entity interface{}) {
	tgGroup.TgGroupEntity, _ = entity.(*TgGroupEntity)
}

func (tgGroup *TgGroup) SetIntID(id int64) {
	tgGroup.ID = id
}
