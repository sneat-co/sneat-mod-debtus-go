package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

func NewRewardDalGae() rewardDalGae {
	return rewardDalGae{}
}

type rewardDalGae struct {
}

var _ dal.RewardDal = (*rewardDalGae)(nil)

func (_ rewardDalGae) InsertReward(c context.Context, rewardEntity *models.RewardEntity) (reward models.Reward, err error) {
	reward.RewardEntity = rewardEntity
	return reward, dal.DB.Update(c, &reward)
}
