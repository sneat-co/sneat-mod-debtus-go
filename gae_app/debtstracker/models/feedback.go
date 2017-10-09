package models

import (
	"time"
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"github.com/strongo/app/db"
)

const FeedbackKind = "Feedback"

type FeedbackEntity struct {
	general.CreatedOn
	UserID  int64
	Created time.Time
	Rate    string
	Text    string `datastore:",noindex"`
}

type Feedback struct {
	db.NoStrID
	ID int64
	*FeedbackEntity
}

func (o *Feedback) Kind() string {
	return FeedbackKind
}

func (o *Feedback) IntID() int64 {
	return o.ID
}

func (o *Feedback) Entity() interface{} {
	return o.FeedbackEntity
}

func (o *Feedback) SetEntity(entity interface{}) {
	o.FeedbackEntity = entity.(*FeedbackEntity)
}

func (o *Feedback) SetIntID(id int64) {
	o.ID = id
}
