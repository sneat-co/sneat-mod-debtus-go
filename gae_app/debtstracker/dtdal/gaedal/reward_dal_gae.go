package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

func NewRewardDalGae() rewardDalGae {
	return rewardDalGae{}
}

type rewardDalGae struct {
}

var _ dtdal.RewardDal = (*rewardDalGae)(nil)

func (rewardDalGae) InsertReward(c context.Context, rewardEntity *models.RewardEntity) (reward models.Reward, err error) {
	reward.RewardEntity = rewardEntity
	return reward, dtdal.DB.Update(c, &reward)
}
