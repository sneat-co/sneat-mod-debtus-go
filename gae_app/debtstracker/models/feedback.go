package models

import (
	"github.com/strongo/dalgo/record"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/general"
)

const FeedbackKind = "Feedback"

type FeedbackEntity struct {
	general.CreatedOn
	UserID  int64
	Created time.Time
	Rate    string
	Text    string `datastore:",noindex"`
}

//var _ db.EntityHolder = (*Feedback)(nil)

type Feedback struct {
	record.WithID[int]
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
