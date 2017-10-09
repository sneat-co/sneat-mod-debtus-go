package models

import (
	"github.com/strongo/app/db"
	"github.com/strongo/app/user"
)

const SplitKind = "Split"

type Split struct {
	db.IntegerID
	db.NoStrID
	*SplitEntity
}

var _ db.EntityHolder = (*Split)(nil)

type SplitEntity struct {
	user.OwnedByUser
	BillIDs []string `datastore:",noindex"`
}

func (Split) Kind() string {
	return SplitKind
}

func (record Split) Entity() interface{} {
	return record.SplitEntity
}

func (record *Split) SetEntity(entity interface{}) {
	record.SplitEntity = entity.(*SplitEntity)
}
