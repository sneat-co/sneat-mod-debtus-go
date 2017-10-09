package models

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"github.com/strongo/app/db"
	"time"
)

type InviteBy string

const (
	InviteByTelegram = InviteBy("telegram")
	InviteByFbm      = InviteBy("fbm")
	InviteByEmail    = InviteBy("email")
	InviteBySms      = InviteBy("sms")
)

const InviteKind = "Invite"

type InviteType string

const (
	InviteTypePersonal = "personal"
	InviteTypePublic   = "public"
)

type InviteEntity struct {
	Channel      string `datastore:",noindex"`
	DtCreated    time.Time
	DtActiveFrom time.Time
	DtActiveTill time.Time
	//
	MaxClaimsCount int32 `datastore:",noindex"`
	ClaimedCount   int32
	LastClaimIDs   []int64 `datastore:",noindex"`
	LastClaimed    time.Time
	//DtClaimed       time.Time
	CreatedByUserID int64
	general.CreatedOn

	Related string

	Type string

	ToName          string `datastore:",noindex"`
	ToEmail         string
	ToEmailOriginal string `datastore:",noindex"`
	ToPhoneNumber   int64
	ToUrl           string
}

type Invite struct {
	db.NoStrID
	ID string
	*InviteEntity
}
