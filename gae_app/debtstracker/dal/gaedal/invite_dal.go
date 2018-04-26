package gaedal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"context"
	"google.golang.org/appengine/datastore"
)

func NewInviteKey(c context.Context, inviteCode string) *datastore.Key {
	return gaedb.NewKey(c, models.InviteKind, inviteCode, 0, nil)
}

type InviteDalGae struct {
}

func NewInviteDalGae() InviteDalGae {
	return InviteDalGae{}
}

func (InviteDalGae) GetInvite(c context.Context, inviteCode string) (*models.InviteEntity, error) {
	inviteKey := gaedb.NewKey(c, models.InviteKind, inviteCode, 0, nil)
	var inviteEntity models.InviteEntity
	err := gaedb.Get(c, inviteKey, &inviteEntity)
	if err == datastore.ErrNoSuchEntity {
		return nil, db.NewErrNotFoundByStrID(models.InviteKind, inviteCode, nil)
	}
	return &inviteEntity, err
}

func (InviteDalGae) ClaimInvite(c context.Context, userID int64, inviteCode, claimedOn, claimedVia string) (err error) {
	err = gaedb.RunInTransaction(c, func(tc context.Context) error {
		inviteKey := gaedb.NewKey(tc, models.InviteKind, inviteCode, 0, nil)
		var invite models.InviteEntity

		if err = gaedb.Get(tc, inviteKey, &invite); err == nil {
			log.Debugf(c, "Invite found")
			// TODO: Check invite.For
			inviteClaim := NewInviteClaim(inviteCode, userID, claimedOn, claimedVia)
			//invite.ClaimedCount += 1
			inviteClaimKey := NewInviteClaimIncompleteKey(c)
			userKey := gaedb.NewKey(tc, models.AppUserKind, "", userID, nil)
			user := new(models.AppUserEntity)
			if err = gaedb.Get(tc, userKey, user); err != nil {
				return err
			}
			user.InvitedByUserID = invite.CreatedByUserID

			keysToPut := []*datastore.Key{inviteClaimKey, userKey}
			entitiesToPut := []interface{}{inviteClaim, user}

			if keysToPut, err := gaedb.PutMulti(tc, keysToPut, entitiesToPut); err != nil {
				return errors.Wrapf(err, "Failed to save %v entities (%v)", len(entitiesToPut), keysToPut)
			}
			log.Debugf(c, "inviteClaimKey.IntegerID(): %v, returned.IntegerID(): %v", inviteClaimKey.IntID(), keysToPut[0].IntID())
			inviteClaimKey = keysToPut[0]
			DelayUpdateInviteClaimedCount(tc, inviteClaimKey.IntID())

			return err
		} else if err == datastore.ErrNoSuchEntity {
			return db.NewErrNotFoundByStrID(models.InviteKind, inviteCode, err)
		}
		return err
	}, &datastore.TransactionOptions{XG: true})
	return
}

const (
	AUTO_GENERATE_INVITE_CODE = ""
	INVITE_CODE_LENGTH        = 5
	PERSONAL_INVITE           = 1
)

func (InviteDalGae) CreatePersonalInvite(ec strongo.ExecutionContext, userID int64, inviteBy models.InviteBy, inviteToAddress, createdOnPlatform, createdOnID, related string) (models.Invite, error) {
	return createInvite(ec, models.InviteTypePersonal, userID, inviteBy, inviteToAddress, createdOnPlatform, createdOnID, INVITE_CODE_LENGTH, AUTO_GENERATE_INVITE_CODE, related, PERSONAL_INVITE)
}

func (InviteDalGae) CreateMassInvite(ec strongo.ExecutionContext, userID int64, inviteCode string, maxClaimsCount int32, createdOnPlatform string) (invite models.Invite, err error) {
	invite, err = createInvite(ec, models.InviteTypePublic, userID, "", "", createdOnPlatform, "", uint8(len(inviteCode)), inviteCode, "", maxClaimsCount)
	return
}

func createInvite(ec strongo.ExecutionContext, inviteType models.InviteType, userID int64, inviteBy models.InviteBy, inviteToAddress, createdOnPlatform, createdOnID string, inviteCodeLen uint8, inviteCode, related string, maxClaimsCount int32) (invite models.Invite, err error) {
	if inviteCode != AUTO_GENERATE_INVITE_CODE && !dal.InviteCodeRegex.Match([]byte(inviteCode)) {
		err = fmt.Errorf("Invalid invite code: %v", inviteCode)
		return
	}
	if related != "" && len(strings.Split(related, "=")) != 2 {
		panic(fmt.Sprintf("Invalid format for related: %v", related))
	}
	c := ec.Context()

	dtCreated := time.Now()
	inviteEntity := models.InviteEntity{
		Type:    string(inviteType),
		Channel: string(inviteBy),
		CreatedOn: general.CreatedOn{
			CreatedOnPlatform: createdOnPlatform,
			CreatedOnID:       createdOnID,
		},
		DtCreated:       dtCreated,
		CreatedByUserID: userID,
		Related:         related,
		MaxClaimsCount:  maxClaimsCount,
		DtActiveFrom:    dtCreated,
		DtActiveTill:    dtCreated.AddDate(100, 0, 0), // By default is active for 100 years
	}
	invite.InviteEntity = &inviteEntity
	switch inviteBy {
	case models.InviteByEmail:
		if inviteToAddress == "" {
			panic("Emmail address is not supplied")
		}
		if strings.Index(inviteToAddress, "@") <= 0 || strings.Index(inviteToAddress, ".") <= 0 {
			panic("Invalid email address")
		}
		inviteEntity.ToEmail = strings.ToLower(inviteToAddress)
		if inviteToAddress != strings.ToLower(inviteToAddress) {
			invite.ToEmailOriginal = inviteToAddress
		}
	case models.InviteBySms:
		var phoneNumber int64
		phoneNumber, err = strconv.ParseInt(inviteToAddress, 10, 64)
		if err != nil {
			return
		}
		inviteEntity.ToPhoneNumber = phoneNumber
	}
	err = gaedb.RunInTransaction(c, func(tc context.Context) error {
		var inviteKey *datastore.Key
		if inviteCode != AUTO_GENERATE_INVITE_CODE {
			inviteKey = gaedb.NewKey(c, models.InviteKind, inviteCode, 0, nil)
		} else {
			for {
				if inviteCodeLen == 0 {
					inviteCodeLen = INVITE_CODE_LENGTH
				}
				inviteCode = dal.RandomCode(inviteCodeLen)
				inviteKey = NewInviteKey(tc, inviteCode)
				var existingInvite models.InviteEntity
				err := gaedb.Get(c, inviteKey, existingInvite)
				if err == datastore.ErrNoSuchEntity {
					log.Debugf(c, "New invite code: %v", inviteCode)
					break
				} else {
					log.Warningf(c, "Already existign invide code: %v", inviteCode)
				}
			}
		}
		inviteKey, err = gaedb.Put(c, inviteKey, &invite)
		return err
	}, nil)
	if err == nil {
		log.Infof(c, "Invite created with code: %v", inviteCode)
	} else {
		log.Errorf(c, "Failed to create invite with code: %v", err)
	}
	return
}

func (InviteDalGae) ClaimInvite2(c context.Context, inviteCode string, inviteEntity *models.InviteEntity, claimedByUserID int64, claimedOn, claimedVia string) (invite models.Invite, err error) {
	var inviteClaimKey *datastore.Key
	err = gaedb.RunInTransaction(c, func(tc context.Context) error {
		inviteKey := NewInviteKey(tc, inviteCode)
		userKey := NewAppUserKey(tc, claimedByUserID)
		var user models.AppUserEntity
		if err = gaedb.GetMulti(tc, []*datastore.Key{inviteKey, userKey}, []interface{}{inviteEntity, &user}); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("Failed to get entities by keys (%v)", []*datastore.Key{inviteKey, userKey}))
		}

		inviteEntity.ClaimedCount += 1
		if inviteEntity.MaxClaimsCount > 0 && inviteEntity.ClaimedCount > inviteEntity.MaxClaimsCount {
			return fmt.Errorf("invite.ClaimedCount > invite.MaxClaimsCount: %v > %v", inviteEntity.ClaimedCount, inviteEntity.MaxClaimsCount)
		}
		inviteClaimKey = NewInviteClaimIncompleteKey(c)
		inviteClaim := NewInviteClaim(inviteCode, claimedByUserID, claimedOn, claimedVia)
		keys := []*datastore.Key{inviteClaimKey, inviteKey}
		entities := []interface{}{inviteClaim, inviteEntity}

		userChanged := updateUserContactDetails(&user, inviteEntity)

		if user.DtAccessGranted.IsZero() {
			user.DtAccessGranted = time.Now()
			userChanged = true
		}
		if inviteEntity.MaxClaimsCount == 1 {
			user.InvitedByUserID = inviteEntity.CreatedByUserID
			userChanged = true
			counterpartyQuery := datastore.NewQuery(models.ContactKind)
			counterpartyQuery = counterpartyQuery.Filter("UserID =", claimedByUserID)
			counterpartyQuery = counterpartyQuery.Filter("CounterpartyUserID =", inviteEntity.CreatedByUserID)
			var counterparties []*models.ContactEntity
			counterpartiesKeys, err := counterpartyQuery.Limit(1).GetAll(c, &counterparties) // Use out-of-transaction context
			if err != nil {
				return errors.Wrap(err, "Failed to load counterparty by CounterpartyUserID")
			}
			if len(counterpartiesKeys) == 0 {
				counterpartyKey := NewContactIncompleteKey(tc)
				inviteCreator, err := dal.User.GetUserByID(c, inviteEntity.CreatedByUserID)
				if err != nil {
					return errors.Wrap(err, "Failed to get invite creator user")
				}

				counterparty := models.NewContactEntity(claimedByUserID, models.ContactDetails{
					FirstName:    inviteCreator.FirstName,
					LastName:     inviteCreator.LastName,
					Username:     inviteCreator.Username,
					EmailContact: inviteCreator.EmailContact,
					PhoneContact: inviteCreator.PhoneContact,
				})

				keys = append(keys, counterpartyKey)
				entities = append(entities, counterparty)
			}
		}

		if userChanged {
			keys = append(keys, userKey)
			entities = append(entities, &user)
		}

		keys, err = gaedb.PutMulti(tc, keys, entities)
		if err != nil {
			err = errors.Wrapf(err, "Failed to put %v entities", len(keys))
			return err
		}
		inviteClaimKey = keys[0]

		return err
	}, &datastore.TransactionOptions{XG: true})
	if err != nil {
		return
	}
	invite = models.Invite{ID: inviteClaimKey.StringID(), InviteEntity: inviteEntity}
	return
}

func updateUserContactDetails(user *models.AppUserEntity, invite *models.InviteEntity) (changed bool) {
	switch models.InviteBy(invite.Channel) {
	case models.InviteByEmail:
		changed = !user.EmailConfirmed
		user.SetEmail(invite.ToEmail, true)
	case models.InviteBySms:
		if invite.ToPhoneNumber != 0 {
			changed = !user.PhoneNumberConfirmed
			user.PhoneNumber = invite.ToPhoneNumber
			user.PhoneNumberConfirmed = true
		}
	}
	return
}
