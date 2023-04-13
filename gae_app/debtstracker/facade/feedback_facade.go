package facade

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/log"
)

func SaveFeedback(c context.Context, tx dal.ReadwriteTransaction, feedbackID int64, feedbackEntity *models.FeedbackData) (feedback models.Feedback, user models.AppUser, err error) {
	if c == nil {
		panic("c == nil")
	}
	log.Debugf(c, "FeedbackDalGae.SaveFeedback(feedbackEntity:%v)", feedbackEntity)
	if feedbackEntity == nil {
		panic("feedbackEntity == nil")
	}
	if feedbackEntity.UserID == 0 {
		panic("feedbackEntity.UserID == 0")
	}
	if feedbackEntity.Rate == "" {
		panic("feedbackEntity.Rate is empty string")
	}
	feedback = models.Feedback{FeedbackData: feedbackEntity}
	user, err = User.GetUserByID(c, tx, feedbackEntity.UserID)
	if err != nil {
		err = fmt.Errorf("failed to get user by ID=%d: %w", feedbackEntity.UserID, err)
	}
	user.Data.LastFeedbackRate = feedbackEntity.Rate
	if feedbackEntity.Created.IsZero() {
		now := time.Now()
		user.Data.LastFeedbackAt = now
		feedbackEntity.Created = now
	} else {
		user.Data.LastTransferAt = feedbackEntity.Created
	}
	if err = tx.SetMulti(c, []dal.Record{feedback.Record, user.Record}); err != nil {
		err = fmt.Errorf("failed to put feedback & user entities to datastore: %w", err)
	}
	return
}
