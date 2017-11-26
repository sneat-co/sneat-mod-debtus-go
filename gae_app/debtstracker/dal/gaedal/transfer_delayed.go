package gaedal

import (
	"fmt"
	"sync"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
)

func (_ TransferDalGae) DelayUpdateTransfersWithCounterparty(c context.Context, creatorCounterpartyID, counterpartyCounterpartyID int64) error {
	log.Debugf(c, "DelayUpdateTransfersWithCounterparty(creatorCounterpartyID=%d, counterpartyCounterpartyID=%d)", creatorCounterpartyID, counterpartyCounterpartyID)
	if creatorCounterpartyID == 0 {
		return errors.New("creatorCounterpartyID == 0")
	}
	if counterpartyCounterpartyID == 0 {
		return errors.New("counterpartyCounterpartyID == 0")
	}
	if task, err := gae.CreateDelayTask(common.QUEUE_TRANSFERS, DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY, delayedUpdateTransfersWithCounterparty, creatorCounterpartyID, counterpartyCounterpartyID); err != nil {
		return err
	} else if task, err = gae.AddTaskToQueue(c, task, common.QUEUE_TRANSFERS); err != nil {
		return err
	}
	return nil
}

const (
	DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY  = "update-transfers-with-counterparty"
	DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY = "update-1-transfer-with-counterparty"
)

var delayedUpdateTransfersWithCounterparty = delay.Func(DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY, func(c context.Context, creatorCounterpartyID, counterpartyCounterpartyID int64) error {
	log.Infof(c, "delayedUpdateTransfersWithCounterparty(creatorCounterpartyID=%d, counterpartyCounterpartyID=%d)", creatorCounterpartyID, counterpartyCounterpartyID)
	if creatorCounterpartyID == 0 {
		log.Errorf(c, "creatorCounterpartyID == 0")
		return nil
	}
	if counterpartyCounterpartyID == 0 {
		log.Errorf(c, "counterpartyCounterpartyID == 0")
		return nil
	}
	query := datastore.NewQuery(models.TransferKind).KeysOnly()
	query = query.Filter("BothCounterpartyIDs =", creatorCounterpartyID).Filter("BothCounterpartyIDs =", 0)
	query = query.Order("-DtCreated") // We don't need order here but it would be nice to update recent first and we have index in place anyway
	var transfers []*models.TransferEntity
	if keys, err := query.GetAll(c, transfers); err != nil {
		return errors.Wrap(err, "Failed to load transfers")
	} else if len(keys) > 0 {
		log.Infof(c, "Loaded %v keys: %v", len(keys), keys)
		delayDuration := 10 * time.Microsecond
		for _, key := range keys {
			if task, err := gae.CreateDelayTask(common.QUEUE_TRANSFERS, DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY, delayedUpdateTransferWithCounterparty, key.IntID(), counterpartyCounterpartyID); err != nil {
				return errors.Wrapf(err, "Failed to create task for transfer id=%v", key.IntID())
			} else {
				task.Delay = delayDuration
				delayDuration += 10 * time.Microsecond
				if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_TRANSFERS); err != nil {
					return errors.Wrapf(err, "Failed to add task for transfer %v to queue [%v]: %v", key, common.QUEUE_TRANSFERS)
				}
			}
		}
	} else {
		query := datastore.NewQuery(models.TransferKind).KeysOnly()
		query = query.Filter("BothCounterpartyIDs =", creatorCounterpartyID).Filter("BothCounterpartyIDs =", counterpartyCounterpartyID)
		query = query.Limit(1).KeysOnly()
		if keys, err := query.GetAll(c, nil); err != nil {
			return errors.Wrap(err, "Failed to load transfers by 2 counterparty IDs")
		} else if len(keys) > 0 {
			log.Infof(c, "No transfers found to update counterparty details")
		} else {
			log.Warningf(c, "No transfers found to update counterparty details")
		}
	}
	return nil
})

var delayedUpdateTransferWithCounterparty = delay.Func(DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY,
	func(c context.Context, transferID, counterpartyCounterpartyID int64) error {
		log.Debugf(c, "delayedUpdateTransferWithCounterparty(transferID=%d, counterpartyCounterpartyID=%d)", transferID, counterpartyCounterpartyID)
		if transferID == 0 {
			log.Errorf(c, "transferID == 0")
			return nil
		}
		if counterpartyCounterpartyID == 0 {
			log.Errorf(c, "counterpartyCounterpartyID == 0")
			return nil
		}

		counterpartyCounterparty, err := dal.Contact.GetContactByID(c, counterpartyCounterpartyID)
		if err != nil {
			log.Errorf(c, err.Error())
			if db.IsNotFound(err) {
				return nil
			}
			return err
		}

		log.Debugf(c, "counterpartyCounterparty: %v", counterpartyCounterparty)

		counterpartyUser, err := dal.User.GetUserByID(c, counterpartyCounterparty.UserID)
		if err != nil {
			log.Errorf(c, err.Error())
			if db.IsNotFound(err) {
				return nil
			}
			return err
		}

		log.Debugf(c, "counterpartyUser: %v", *counterpartyUser.AppUserEntity)

		if err := dal.DB.RunInTransaction(c, func(tc context.Context) error {
			transfer, err := dal.Transfer.GetTransferByID(tc, transferID)
			if err != nil {
				return err
			}
			changed := false

			// TODO: allow to pass creator counterparty as well. Match by userID

			log.Debugf(c, "transfer.From() before: %v", transfer.From())
			log.Debugf(c, "transfer.To() before: %v", transfer.To())

			// Update transfer creator
			{
				transferCreator := transfer.Creator()
				log.Debugf(c, "transferCreator before: %v", transferCreator)
				if transferCreator.ContactID == 0 {
					transferCreator.ContactID = counterpartyCounterparty.ID
					changed = true
				} else if transferCreator.ContactID != counterpartyCounterparty.ID {
					err = fmt.Errorf("transferCounterparty.ContactID != counterpartyCounterparty.ID: %d != %d", transferCreator.ContactID, counterpartyCounterparty.ID)
					return err
				} else {
					log.Debugf(c, "transferCounterparty.ContactID == counterpartyCounterparty.ID: %d", transferCreator.ContactID)
				}
				if transferCreator.ContactName == "" || transferCreator.ContactName != counterpartyCounterparty.FullName() {
					transferCreator.ContactName = counterpartyCounterparty.FullName()
					changed = true
				}
				log.Debugf(c, "transferCreator after: %v", transferCreator)
				log.Debugf(c, "transfer.Creator() after: %v", transfer.Creator())
			}

			// Update transfer counterparty
			{
				transferCounterparty := transfer.Counterparty()
				log.Debugf(c, "transferCounterparty before: %v", transferCounterparty)
				if transferCounterparty.UserID == 0 {
					transferCounterparty.UserID = counterpartyCounterparty.UserID
					changed = true
				} else if transferCounterparty.UserID != counterpartyCounterparty.UserID {
					err = fmt.Errorf("transferCounterparty.UserID != counterpartyCounterparty.UserID: %d != %d", transferCounterparty.UserID, counterpartyCounterparty.UserID)
					return err
				} else {
					log.Debugf(c, "transferCounterparty.UserID == counterpartyCounterparty.UserID: %d", transferCounterparty.UserID)
				}
				if transferCounterparty.UserName == "" || transferCounterparty.UserName != counterpartyUser.FullName() {
					transferCounterparty.UserName = counterpartyUser.FullName()
					changed = true
				}
				log.Debugf(c, "transferCounterparty after: %v", transferCounterparty)
				log.Debugf(c, "transfer.Contact() after: %v", transfer.Counterparty())
			}
			log.Debugf(c, "transfer.From() after: %v", transfer.From())
			log.Debugf(c, "transfer.To() after: %v", transfer.To())

			if changed {
				if err = dal.Transfer.SaveTransfer(tc, transfer); err != nil {
					return err
				}
				if !transfer.DtDueOn.IsZero() {
					var counterpartyUser models.AppUser
					if counterpartyUser, err = dal.User.GetUserByID(c, counterpartyCounterparty.UserID); err != nil {
						return err
					}

					if !counterpartyUser.HasDueTransfers {
						if err = dal.User.DelayUpdateUserHasDueTransfers(tc, counterpartyCounterparty.UserID); err != nil {
							return err
						}
					}
				}
				log.Infof(c, "Transfer saved to datastore")
				return nil
			} else {
				log.Infof(c, "No chanes for the transfer")
			}
			return nil
		}, nil); err != nil {
			panic(fmt.Sprintf("Failed to update transfer (%d): %v", transferID, err.Error()))
		} else {
			log.Infof(c, "Transaction succesfully completed")
		}
		return nil
	})

const (
	UPDATE_TRANSFERS_WITH_CREATOR_NAME = "update-transfers-with-creator-name"
)

func DelayUpdateTransfersWithCreatorName(c context.Context, userID int64) error {
	return gae.CallDelayFunc(c, common.QUEUE_TRANSFERS, UPDATE_TRANSFERS_WITH_CREATOR_NAME, delayedUpdateTransfersWithCreatorName, userID)
}

var delayedUpdateTransfersWithCreatorName = delay.Func(UPDATE_TRANSFERS_WITH_CREATOR_NAME, func(c context.Context, userID int64) error {
	log.Debugf(c, "delayedUpdateTransfersWithCreatorName(userID=%d)", userID)

	user, err := dal.User.GetUserByID(c, userID)
	if err != nil {
		log.Errorf(c, err.Error())
		if db.IsNotFound(err) {
			err = nil
		}
		return err
	}

	userName := user.FullName()

	query := datastore.NewQuery(models.TransferKind).KeysOnly().Filter("BothUserIDs =", userID)

	t := query.Run(c)
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		var transferEntity models.TransferEntity
		key, err := t.Next(&transferEntity)
		if err != nil {
			if err == datastore.Done {
				return nil
			}
			log.Errorf(c, err.Error())
			return err
		}
		wg.Add(1)
		go func(transferID int64) {
			defer wg.Done()
			err := dal.DB.RunInTransaction(c, func(c context.Context) error {
				transfer, err := dal.Transfer.GetTransferByID(c, transferID)
				if err != nil {
					return err
				}
				changed := false
				switch userID {
				case transfer.From().UserID:
					if from := transfer.From(); from.UserName != userName {
						from.UserName = userName
						changed = true
					}
				case transfer.To().UserID:
					if to := transfer.To(); to.UserName != userName {
						to.UserName = userName
						changed = true
					}
				default:
					log.Infof(c, "Transfer(%d) creator is not a counterparty")
				}
				if changed {
					if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
						return err
					}
				}
				return err
			}, dal.SingleGroupTransaction)
			if err != nil {
				log.Errorf(c, err.Error())
			}
			return
		}(key.IntID())
	}
})
