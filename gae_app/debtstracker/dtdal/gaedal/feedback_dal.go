package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
)

type FeedbackDalGae struct {
}

func NewFeedbackDalGae() FeedbackDalGae {
	return FeedbackDalGae{}
}

func (FeedbackDalGae) GetFeedbackByID(c context.Context, tx dal.ReadSession, feedbackID int64) (feedback models.Feedback, err error) {
	if tx == nil {
		if tx, err = facade.GetDatabase(c); err != nil {
			return
		}
	}
	feedback = models.NewFeedback(feedbackID, nil)
	return feedback, tx.Get(c, feedback.Record)
}
