package models4debtus

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/general4debtus"
	"time"
)

type InviteBy string

const (
	InviteByTelegram       = InviteBy("telegram")
	InviteByFbm            = InviteBy("fbm")
	InviteByEmail          = InviteBy("email")
	InviteBySms            = InviteBy("sms")
	InviteByLinkToTelegram = InviteBy("link2tg")
)

type InviteType string

const (
	InviteTypePersonal = "personal"
	InviteTypePublic   = "public"
)

const InviteKind = "Invite"

type Invite = record.DataWithID[string, *InviteData]

func NewInviteKey(inviteCode string) *dal.Key {
	return dal.NewKeyWithID(InviteKind, inviteCode)
}

func NewInvite(id string, data *InviteData) Invite {
	key := NewInviteKey(id)
	return Invite{
		WithID: record.NewWithID(id, key, &data),
		Data:   data,
	}
}

type InviteData struct {
	Channel      string `firestore:",omitempty"`
	DtCreated    time.Time
	DtActiveFrom time.Time
	DtActiveTill time.Time
	//
	MaxClaimsCount int32 `firestore:",omitempty"`
	ClaimedCount   int32
	LastClaimIDs   []int64 `firestore:",omitempty"`
	LastClaimed    time.Time
	//DtClaimed       time.Time
	CreatedByUserID string
	general4debtus.CreatedOn

	Related string

	Type string

	ToName          string `firestore:",omitempty"`
	ToEmail         string
	ToEmailOriginal string `firestore:",omitempty"`
	ToPhoneNumber   int64
	ToUrl           string
}
