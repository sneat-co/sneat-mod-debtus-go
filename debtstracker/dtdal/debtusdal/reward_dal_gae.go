package debtusdal

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
)

func NewRewardDalGae() rewardDalGae {
	return rewardDalGae{}
}

type rewardDalGae struct {
}

var _ dtdal.RewardDal = (*rewardDalGae)(nil)

func (rewardDalGae) InsertReward(ctx context.Context, tx dal.ReadwriteTransaction, rewardEntity *models4debtus.RewardDbo) (reward models4debtus.Reward, err error) {
	reward = models4debtus.NewRewardWithIncompleteKey(nil)
	if err = tx.Insert(ctx, reward.Record); err != nil {
		return
	}
	reward.ID = reward.Record.Key().ID.(string)
	return
}
