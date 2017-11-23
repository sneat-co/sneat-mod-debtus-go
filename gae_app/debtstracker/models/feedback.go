package models

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"github.com/strongo/db"
	"time"
)

const FeedbackKind = "Feedback"

type FeedbackEntity struct {
	general.CreatedOn
	UserID  int64
	Created time.Time
	Rate    string
	Text    string `datastore:",noindex"`
}

var _ db.EntityHolder = (*Feedback)(nil)

type Feedback struct {
	db.IntegerID
	*FeedbackEntity
}

func (o *Feedback) Kind() string {
	return FeedbackKind
}

func (o Feedback) Entity() interface{} {
	return o.FeedbackEntity
}

func (Feedback) NewEntity() interface{} {
	return new(FeedbackEntity)
}

func (o *Feedback) SetEntity(entity interface{}) {
	o.FeedbackEntity = entity.(*FeedbackEntity)
}
