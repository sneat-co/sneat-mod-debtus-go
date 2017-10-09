package gaedal

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
	"github.com/strongo/app/gaedb"
)

func NewInviteClaimIncompleteKey(c context.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, models.InviteClaimKind, nil)
}

func NewInviteClaimKey(c context.Context, claimID int64) *datastore.Key {
	return gaedb.NewKey(c, models.InviteClaimKind, "", claimID, nil)
}

func NewInviteClaim(inviteCode string, userID int64, claimedOn, claimedVia string) *models.InviteClaim {
	return &models.InviteClaim{
		InviteCode: inviteCode,
		UserID:     userID,
		ClaimedOn:  claimedOn,
		ClaimedVia: claimedVia,
		DtClaimed:  time.Now(),
	}
}
