package facade

import (
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/log"
)

func SaveFeedback(c context.Context, feedbackID int64, feedbackEntity *models.FeedbackEntity) (feedback models.Feedback, user models.AppUser, err error) {
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
	feedback = models.Feedback{FeedbackEntity: feedbackEntity}
	user, err = User.GetUserByID(c, feedbackEntity.UserID)
	if err != nil {
		err = errors.Wrapf(err, "Failed to get user by ID=%d", feedbackEntity.UserID)
	}
	user.LastFeedbackRate = feedbackEntity.Rate
	if feedbackEntity.Created.IsZero() {
		now := time.Now()
		user.LastFeedbackAt = now
		feedbackEntity.Created = now
	} else {
		user.LastTransferAt = feedbackEntity.Created
	}
	if err = dtdal.DB.UpdateMulti(c, []db.EntityHolder{&feedback, &user}); err != nil {
		err = errors.Wrap(err, "Failed to put feedback & user entities to datastore")
	}
	return
}
