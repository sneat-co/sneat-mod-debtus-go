package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/dal-go/dalgo/dal"
)

func NewRewardDalGae() rewardDalGae {
	return rewardDalGae{}
}

type rewardDalGae struct {
}

var _ dtdal.RewardDal = (*rewardDalGae)(nil)

func (rewardDalGae) InsertReward(c context.Context, tx dal.ReadwriteTransaction, rewardEntity *models.RewardData) (reward models.Reward, err error) {
	reward = models.NewRewardWithIncompleteKey(nil)
	if err = tx.Insert(c, reward.Record); err != nil {
		return
	}
	reward.ID = reward.Record.Key().ID.(int)
	return
}
