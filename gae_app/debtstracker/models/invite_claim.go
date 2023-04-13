package models

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"time"
)

const InviteClaimKind = "InviteClaim"

type InviteClaim struct {
	record.WithID[int64]
	Data *InviteClaimData
}

type InviteClaimData struct {
	InviteCode string // We don't use it as parent key as can be a bottleneck for public invites
	UserID     int64
	DtClaimed  time.Time
	ClaimedOn  string // For example: "Telegram"
	ClaimedVia string // For the Telegram it would be bot name
}

func NewInviteClaimIncompleteKey() *dal.Key {
	return dal.NewKey(InviteClaimKind)
}

func NewInviteClaimKey(claimID int64) *dal.Key {
	return dal.NewKeyWithID(InviteClaimKind, claimID)
}

func NewInviteClaim(inviteCode string, userID int64, claimedOn, claimedVia string) *InviteClaimData {
	return &InviteClaimData{
		InviteCode: inviteCode,
		UserID:     userID,
		ClaimedOn:  claimedOn,
		ClaimedVia: claimedVia,
		DtClaimed:  time.Now(),
	}
}
