package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

func NewRewardDalGae() rewardDalGae {
	return rewardDalGae{}
}

type rewardDalGae struct {
}

var _ dal.RewardDal = (*rewardDalGae)(nil)

func (rewardDalGae) InsertReward(c context.Context, rewardEntity *models.RewardEntity) (reward models.Reward, err error) {
	reward.RewardEntity = rewardEntity
	return reward, dal.DB.Update(c, &reward)
}
