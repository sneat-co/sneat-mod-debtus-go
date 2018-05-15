package models

import (
	"github.com/strongo/app/user"
	"github.com/strongo/db"
)

const SplitKind = "Split"

type Split struct {
	db.IntegerID
	*SplitEntity
}

var _ db.EntityHolder = (*Split)(nil)

type SplitEntity struct {
	user.OwnedByUserWithIntID
	BillIDs []string `datastore:",noindex"`
}

func (Split) Kind() string {
	return SplitKind
}

func (record Split) Entity() interface{} {
	return record.SplitEntity
}

func (Split) NewEntity() interface{} {
	return new(SplitEntity)
}

func (record *Split) SetEntity(entity interface{}) {
	if entity == nil {
		record.SplitEntity = nil
	} else {
		record.SplitEntity = entity.(*SplitEntity)
	}

}
