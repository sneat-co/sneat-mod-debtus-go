package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/v2/datastore"
)

type FeedbackDalGae struct {
}

func NewFeedbackDalGae() FeedbackDalGae {
	return FeedbackDalGae{}
}

func NewFeedbackKey(c context.Context, feedbackID int64) *datastore.Key {
	return datastore.NewKey(c, models.FeedbackKind, "", feedbackID, nil)
}

func (FeedbackDalGae) GetFeedbackByID(c context.Context, feedbackID int64) (feedback models.Feedback, err error) {
	var entity models.FeedbackEntity
	feedback.ID = feedbackID
	if err = gaedb.Get(c, NewFeedbackKey(c, feedbackID), &entity); err != nil {
		return
	}
	feedback.FeedbackEntity = &entity
	return
}
