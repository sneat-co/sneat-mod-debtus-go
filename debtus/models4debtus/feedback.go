package models4debtus

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/general4debtus"
	"time"
)

const FeedbackKind = "Feedback"

type FeedbackData struct {
	general4debtus.CreatedOn
	// Deprecated: use UserStrID instead
	UserID    int64
	UserStrID string
	Created   time.Time
	Rate      string
	Text      string `firestore:",omitempty"`
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
