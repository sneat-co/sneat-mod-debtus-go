package facade

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"time"
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
	user, err = dal.User.GetUserByID(c, feedbackEntity.UserID)
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
	if err = dal.DB.UpdateMulti(c, []db.EntityHolder{&feedback, &user}); err != nil {
		err = errors.Wrap(err, "Failed to put feedback & user entities to datastore")
	}
	return
}
