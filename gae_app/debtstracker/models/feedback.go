package models

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/general"
)

const FeedbackKind = "Feedback"

type FeedbackData struct {
	general.CreatedOn
	UserID  int64
	Created time.Time
	Rate    string
	Text    string `datastore:",noindex"`
}

//var _ db.EntityHolder = (*Feedback)(nil)

type Feedback struct {
	record.WithID[int64]
	*FeedbackData
}

func NewFeedbackKey(feedbackID int64) *dal.Key {
	return dal.NewKeyWithID(FeedbackKind, feedbackID)
}

func NewFeedback(id int64, data *FeedbackData) Feedback {
	key := NewFeedbackKey(id)
	return Feedback{
		WithID:       record.NewWithID(id, key, &data),
		FeedbackData: data,
	}
}

//func (o *Feedback) Kind() string {
//	return FeedbackKind
//}
//
//func (o Feedback) Entity() interface{} {
//	return o.FeedbackData
//}
//
//func (Feedback) NewEntity() interface{} {
//	return new(FeedbackData)
//}
//
//func (o *Feedback) SetEntity(entity interface{}) {
//	o.FeedbackData = entity.(*FeedbackData)
//}
