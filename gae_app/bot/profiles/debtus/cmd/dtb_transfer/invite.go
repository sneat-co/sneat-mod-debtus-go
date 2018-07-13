package dtb_transfer

//
//import (
//	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal/gaedal"
//	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
//	"bitbucket.org/asterus/debtstracker-server/gae_app/invites"
//	"github.com/pkg/errors"
//	"github.com/strongo/nds"
//	"github.com/strongo/bots-framework/core"
//	"context"
//	"time"
//	"github.com/strongo/app"
//)
//
//func ClaimInviteOnTransfer(ec strongo.ExecutionContext, inviteCode string, invite *invites.Invite) (transferID int64, transfer *models.Transfer, counterpartyID int64, counterparty *models.Contact, err error) {
//	c := ec.Context()
//
//	//if transferID, err = invite.RelatedIntID(); err != nil {
//	//	err = fmt.Errorf("Transfer ID in invite.Related is not an int64: [%v]", invite.Related)
//	//	return
//	//}
//
//	transfer = new(models.Transfer)
//	userID := whc.AppUserIntID()
//	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
//		inviteKey := datastore.NewKey(tc, invites.InviteKind, inviteCode, 0, nil)
//		invite := new(invites.Invite)
//		err = nds.Get(tc, inviteKey, invite)
//		if err != nil {
//			return errors.Wrapf(err, "Failed to get invite by StringID='%v'", inviteKey.StringID())
//		}
//		if invite.CreatedByUserID == userID {
//			err = errors.New("invite.CreatedByUserID == userID")
//			return err
//		}
//		inviteClaimKey := datastore.NewIncompleteKey(tc, invites.InviteClaimKind, nil)
//		inviteClaim := invites.NewInviteClaim(inviteKey.StringID(), userID, whc.BotPlatform().ID(), whc.GetBotCode())
//		invite.ClaimedCount += 1 // ToDo: Can be a bottleneck if invite.MaxClaimsCount > 1 and big
//		userKey := gaedal.NewAppUserKey(tc, userID)
//		transferKey := gaedal.NewTransferKey(tc, transferID)
//		if err = nds.Get(tc, transferKey, transfer); err != nil {
//			return errors.Wrapf(err, "Failed to get transfer by ID=%v", transferKey)
//
//		}
//		if transfer.CreatorUserID == userID {
//			//m = whc.NewMessage("This is your own transfer")
//			err = errors.New("This is your own transfer") // TODO: What do we do?
//			return err
//		}
//		if transfer.Contact().UserID == 0 {
//			user := new(models.AppUser)
//			if err = nds.Get(tc, userKey, user); err != nil {
//				return errors.Wrapf(err, "Failed to get user by ID=%v", userKey.IntID())
//			}
//			if user.InvitedByUserID == 0 {
//				user.InvitedByUserID = invite.CreatedByUserID
//			}
//
//			invites.updateUserContactDetails(user, invite)
//			keysToPut := []*datastore.Key{transferKey, userKey, inviteKey, inviteClaimKey}
//			entitiesToPut := []interface{}{transfer, user, invite, inviteClaim}
//
//			var updateTransferWithCounterpartyDetails = func(counterpartyCounterpartyID int64, counterpartyKey *datastore.Key, counterparty *models.Contact) {
//				log.Debugf(c, "updateTransferWithCounterpartyDetails(counterpartyCounterpartyID=%v)", counterpartyCounterpartyID)
//				counterpartyID = counterpartyCounterpartyID
//				transfer.CounterpartyCounterparty().CounterpartyID = counterpartyCounterpartyID
//				transfer.CounterpartyCounterparty().CounterpartyName = counterparty.GetFullName()
//				//if inlineMessageID != "" {
//				//	transfer.CounterpartyTgReceiptInlineMessageID = inlineMessageID
//				//}
//				transferAmount := transfer.GetAmount()
//				transferVal := transferAmount.Value
//				if transfer.Direction == models.TransferDirectionUser2Counterparty {
//					transferVal *= -1
//				}
//				//user.Add2Balance(transferAmount.Currency, transferVal)
//				//counterparty.Add2Balance(transferAmount.Currency, transferVal)
//				keysToPut = append(keysToPut, counterpartyKey)
//				entitiesToPut = append(entitiesToPut, counterparty)
//			}
//			transfer.Contact().UserID = userID
//			if transfer.CounterpartyCounterparty().CounterpartyID != 0 {
//				// Cleaning just in case
//				transfer.CounterpartyCounterparty().CounterpartyID = 0
//				log.Warningf(c, "Transfer %v had CounterpartyCounterpartyID != 0", transferID)
//			}
//			transferCreatorUser := new(models.AppUser)
//			if transferCreatorUser, err = facade.User.GetUserByID(tc, transfer.CreatorUserID); err != nil {
//				log.Errorf(c, "Failed to load transferCreatorUser by ID (%v): %err", transfer.CreatorUserID, err)
//				return err
//			}
//			creatorCounterpartyKey, creatorCounterparty, err := gaedal.GetCounterpartyByID(tc, transfer.Contact().CounterpartyID)
//			if err != nil {
//				return errors.Wrapf(err, "Failed to call GetCounterpartyByID(%v)", transfer.Contact().CounterpartyID)
//			}
//
//			if user.CounterpartiesCount == 0 {
//				var counterpartKey *datastore.Key
//				counterpartKey, counterparty, err = gaedal.CreateCounterpartyWithinTransaction(tc,
//					userID, transfer.CreatorUserID, transfer.Contact().CounterpartyID, transferCreatorUser.ContactDetails)
//				counterparty.CountOfTransfers = creatorCounterparty.CountOfTransfers
//				if err != nil {
//					return errors.Wrapf(err, "Failed to call CreateCounterpartyWithinTransaction(userID=%v)", userID)
//				}
//				updateTransferWithCounterpartyDetails(counterpartKey.IntID(), counterpartKey, counterparty)
//			} else {
//				counterpartyKeys, counterparties, err := gaedal.GetCounterpartiesByIDs(tc, user.CounterpartyIDs)
//				if err != nil {
//					return errors.Wrapf(err, "Failed to call GetCounterpartiesByIDs(len(user.CounterpartyIDs):%v)", len(user.CounterpartyIDs))
//				}
//				var counterpartyFound bool
//				var i int
//				for i, counterparty = range counterparties {
//					if counterparty.CounterpartyUserID == transfer.CreatorUserID {
//						counterpartyFound = true
//						if counterparty.FirstName == "" {
//							counterparty.FirstName = transferCreatorUser.FirstName
//						}
//						if counterparty.LastName == "" {
//							counterparty.LastName = transferCreatorUser.LastName
//						}
//						updateTransferWithCounterpartyDetails(counterpartyKeys[i].IntID(), counterpartyKeys[i], counterparty)
//						break
//					}
//				}
//				if !counterpartyFound {
//					log.Infof(c, "Contact not found by userID=%v, len(counterparties)=%v", userID, len(counterparties))
//					counterparty = nil
//				}
//			}
//			if counterparty != nil {
//				user.AddCounterpartyID(counterpartyID)
//			}
//			creatorCounterpartyNamesChanged := false
//			if creatorCounterparty.FirstName == "" {
//				creatorCounterparty.FirstName = user.FirstName
//				creatorCounterpartyNamesChanged = true
//			}
//			if creatorCounterparty.LastName == "" {
//				creatorCounterparty.LastName = user.LastName
//				creatorCounterpartyNamesChanged = true
//			}
//			if creatorCounterpartyNamesChanged {
//				creatorCounterparty.UpdateSearchName()
//			}
//			var creatorCounterpartyBalance money.Balance
//			if creatorCounterpartyBalance, err = creatorCounterparty.Balance(); err != nil {
//				return errors.Wrap(err, "Failed to get creatorCounterparty.Balance()")
//			}
//			for currency, value := range creatorCounterpartyBalance {
//				counterparty.Add2Balance(currency, -1*value)
//				user.Add2Balance(currency, -1*value)
//			}
//
//			switch creatorCounterparty.CounterpartyUserID {
//			case 0:
//				creatorCounterparty.CounterpartyUserID = userID
//				creatorCounterparty.CounterpartyCounterpartyID = counterpartyID
//				keysToPut = append(keysToPut, creatorCounterpartyKey)
//				entitiesToPut = append(entitiesToPut, creatorCounterparty)
//				// TODO: Queue task to update all existing transfers
//				if creatorCounterparty.CountOfTransfers > 1 {
//					if err = delayUpdateTransfersWithCounterparty(tc, transfer.CreatorUserID, transfer.Contact().CounterpartyID, models.TransferCounterpartyInfo{
//						UserID:           userID,
//						CounterpartyID:   counterpartyID,
//						Name: counterparty.GetFullName(),
//					}); err != nil {
//						return errors.Wrap(err, "Failed to queeu delayUpdateTransfersWithCounterparty()")
//					}
//				}
//			case userID:
//				log.Infof(c, "creatorCounterparty.CounterpartyUserID already set")
//			default:
//				log.Warningf(c, "creatorCounterparty.CounterpartyUserID is differnt from current user. creatorCounterparty.CounterpartyUserID: %v, currentUserID: %v", creatorCounterparty.CounterpartyUserID, userID)
//			}
//
//			if _, err = nds.PutMulti(tc, keysToPut, entitiesToPut); err != nil {
//				log.Errorf(c, "Failed to call nds.PutMulti(keysToPut=%v, len(entitiesToPut)=%v)", keysToPut, len(entitiesToPut))
//				return err
//			}
//		}
//		if err = gaedal.DelayUpdateUserHasDueTransfers(tc, transfer.Contact().UserID); err != nil {
//			return err
//		}
//		if transfer.DtDueOn.After(time.Now()) {
//			if err := gaedal.DelayCreateReminderForTransferCounterparty(c, transferID); err != nil {
//				return errors.Wrap(err, "Failed to delay creation of reminder for transfer coutnerparty")
//			}
//		} else {
//			if transfer.DtDueOn.IsZero() {
//				log.Debugf(c, "No neeed to create reminder for counterparty as no due date")
//			} else {
//				log.Debugf(c, "No neeed to create reminder for counterparty as due date in past")
//			}
//		}
//		return err
//	}, dal.CrossGroupTransaction)
//	if err != nil {
//		return
//	}
//	log.Debugf(c, "Transaction completed without errors")
//	if err = bots.SetAccessGranted(whc, true); err != nil {
//		err = errors.Wrap(err, "Failed to call bots.SetAccessGranted(whc, true)")
//	}
//	return
//}
