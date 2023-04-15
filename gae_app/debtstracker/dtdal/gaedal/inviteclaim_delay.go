package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
)

func DelayUpdateInviteClaimedCount(c context.Context, claimID int64) error {
	return delayedUpdateInviteClaimedCount.Call(c, claimID)
}

var delayedUpdateInviteClaimedCount = delay.Func("UpdateInviteClaimedCount", func(c context.Context, claimID int64) (err error) {
	log.Debugf(c, "delayedUpdateInviteClaimedCount(claimID=%v)", claimID)
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return err
	}
	err = db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) (err error) {
		claim := models.NewInviteClaim(claimID, nil)
		err = tx.Get(c, claim.Record)
		if err != nil {
			if dal.IsNotFound(err) {
				log.Errorf(c, "Claim not found by id: %v", claimID)
				return nil
			}
			return fmt.Errorf("failed to get InviteClaimData by id=%v: %w", claimID, err)
		}
		invite, err := dtdal.Invite.GetInvite(c, tx, claim.Data.InviteCode)
		if err != nil {
			if dal.IsNotFound(err) {
				log.Errorf(c, "Invite not found by code: %v", claim.Data.InviteCode)
				return nil // Internationally return NIL to avoid retrying
			}
			return err
		}
		for _, cid := range invite.Data.LastClaimIDs {
			if cid == claimID {
				log.Infof(c, "Invite already has been updated for this claim (claimID=%v, inviteCode=%v).", claimID, claim.Data.InviteCode)
				return nil
			}
		}
		invite.Data.ClaimedCount += 1
		if invite.Data.LastClaimed.Before(claim.Data.DtClaimed) {
			invite.Data.LastClaimed = claim.Data.DtClaimed
		}
		invite.Data.LastClaimIDs = append(invite.Data.LastClaimIDs, claimID)
		if len(invite.Data.LastClaimIDs) > 10 {
			invite.Data.LastClaimIDs = invite.Data.LastClaimIDs[len(invite.Data.LastClaimIDs)-10:]
		}

		if err = tx.Set(tc, invite.Record); err != nil {
			return fmt.Errorf("failed to save invite to DB: %w", err)
		}
		return err
	})
	if err != nil {
		log.Errorf(c, "Failed to update Invite.ClaimedCount for claimID=%v", claimID)
	}
	return err
})
