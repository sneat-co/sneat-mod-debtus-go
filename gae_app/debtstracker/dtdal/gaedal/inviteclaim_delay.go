package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
)

func DelayUpdateInviteClaimedCount(c context.Context, claimID int64) error {
	return delayedUpdateInviteClaimedCount.Call(c, claimID)
}

var delayedUpdateInviteClaimedCount = delay.Func("UpdateInviteClaimedCount", func(c context.Context, claimID int64) error {
	log.Debugf(c, "delayedUpdateInviteClaimedCount(claimID=%v)", claimID)
	var claim models.InviteClaim
	err := gaedb.Get(c, NewInviteClaimKey(c, claimID), &claim)
	if err != nil {
		if dal.IsNotFound(err) {
			log.Errorf(c, "Claim not found by id: %v", claimID)
			return nil
		}
		return fmt.Errorf("failed to get InviteClaim by id=%v: %w", claimID, err)
	}
	err = gaedb.RunInTransaction(c, func(tc context.Context) error {
		invite, err := dtdal.Invite.GetInvite(c, claim.InviteCode)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				log.Errorf(c, "Invite not found by code: %v", claim.InviteCode)
				return nil
			}
			return err
		}
		for _, cid := range invite.LastClaimIDs {
			if cid == claimID {
				log.Infof(c, "Invite already has been updated for this claim (claimID=%v, inviteCode=%v).", claimID, claim.InviteCode)
				return nil
			}
		}
		invite.ClaimedCount += 1
		if invite.LastClaimed.Before(claim.DtClaimed) {
			invite.LastClaimed = invite.DtCreated
		}
		invite.LastClaimIDs = append(invite.LastClaimIDs, claimID)
		if len(invite.LastClaimIDs) > 10 {
			invite.LastClaimIDs = invite.LastClaimIDs[len(invite.LastClaimIDs)-10:]
		}
		_, err = gaedb.Put(tc, NewInviteKey(tc, claim.InviteCode), invite)
		if err != nil {
			err = fmt.Errorf("failed to save invite to DB: %w", err)
		}
		return err
	}, nil)
	if err != nil {
		log.Errorf(c, "Failed to update Invite.ClaimedCount for claimID=%v", claimID)
	}
	return err
})
