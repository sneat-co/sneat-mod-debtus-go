package models

import (
	"github.com/strongo/app/db"
	"time"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
)

const RewardKind = "Reward"

type Reward struct {
	db.IntegerID
	db.NoStrID
	*RewardEntity
}

func (_ Reward) Kind() string {
	return RewardKind
}

func (reward Reward) Entity() interface{} {
	if reward.RewardEntity == nil {
		reward.RewardEntity = &RewardEntity{}
	}
	return reward.RewardEntity
}

func (reward *Reward) SetEntity(entity interface{}) {
	reward.RewardEntity = entity.(*RewardEntity)
}

var _ db.EntityHolder = (*Reward)(nil)

type RewardReason string

const (
	RewardReasonInvitedUserJoined         RewardReason = "InvitedUserJoined"
	RewardReasonFriendOfInvitedUserJoined RewardReason = "FriendOfInvitedUserJoined"
)

type RewardEntity struct {
	UserID       int64
	DtCreated    time.Time
	Reason       RewardReason `datastore:",noindex"`
	JoinedUserID int64        `datastore:",noindex"`
	Points       int          `datastore:",noindex"`
}

type UserRewardBalance struct {
	RewardPoints   int
	RewardOptedOut time.Time
	RewardIDs      []int64 `datastore:",noindex"`
}

func (_ UserRewardBalance) cleanProperties(properties []datastore.Property) ([]datastore.Property, error) {
	return gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"RewardPoints": gaedb.IsZeroInt,
		"RewardOptedOut": gaedb.IsZeroTime,
	})
}

func (rewardBalance UserRewardBalance) AddRewardPoints(rewardID int64, rewardPoints int) (changed bool) {
	for _, id := range rewardBalance.RewardIDs {
		if id == rewardID {
			return
		}
	}
	rewardBalance.RewardPoints += rewardPoints
	rewardBalance.RewardIDs = append([]int64{rewardID}, rewardBalance.RewardIDs...)
	return true
}

