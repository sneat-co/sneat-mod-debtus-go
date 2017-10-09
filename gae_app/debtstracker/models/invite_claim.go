package models

import "time"

const InviteClaimKind = "InviteClaim"

type InviteClaim struct {
	InviteCode string // We don't use it as parent key as can be a bottleneck for public invites
	UserID     int64
	DtClaimed  time.Time
	ClaimedOn  string // For example: "Telegram"
	ClaimedVia string // For Telegram it would be bot name
}
