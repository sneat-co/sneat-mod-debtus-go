package gaedal

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"sync"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
	"google.golang.org/appengine/v2/delay"
)

func (TransferDalGae) DelayUpdateTransfersWithCounterparty(c context.Context, creatorCounterpartyID, counterpartyCounterpartyID int64) error {
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
	query = query.Order("-DtCreated") // We don't need order here, but it would be nice to update recent first and we have index in place anyway
	var transfers []*models.TransferEntity
	if keys, err := query.GetAll(c, transfers); err != nil {
		return fmt.Errorf("failed to load transfers: %w", err)
	} else if len(keys) > 0 {
		log.Infof(c, "Loaded %v keys: %v", len(keys), keys)
		delayDuration := 10 * time.Microsecond
		for _, key := range keys {
			if task, err := gae.CreateDelayTask(common.QUEUE_TRANSFERS, DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY, delayedUpdateTransferWithCounterparty, key.IntID(), counterpartyCounterpartyID); err != nil {
				return fmt.Errorf("failed to create task for transfer id=%v: %w", key.IntID(), err)
			} else {
				task.Delay = delayDuration
				delayDuration += 10 * time.Microsecond
				if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_TRANSFERS); err != nil {
					return fmt.Errorf("failed to add task for transfer %v to queue [%v]: %w", key, common.QUEUE_TRANSFERS, err)
				}
			}
		}
	} else {
		query := datastore.NewQuery(models.TransferKind).KeysOnly()
		query = query.Filter("BothCounterpartyIDs =", creatorCounterpartyID).Filter("BothCounterpartyIDs =", counterpartyCounterpartyID)
		query = query.Limit(1).KeysOnly()
		if keys, err := query.GetAll(c, nil); err != nil {
			return fmt.Errorf("failed to load transfers by 2 counterparty IDs: %w", err)
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

		counterpartyCounterparty, err := facade.GetContactByID(c, counterpartyCounterpartyID)
		if err != nil {
			log.Errorf(c, err.Error())
			if dal.IsNotFound(err) {
				return nil
			}
			return err
		}

		log.Debugf(c, "counterpartyCounterparty: %v", counterpartyCounterparty)

		counterpartyUser, err := facade.User.GetUserByID(c, tx, counterpartyCounterparty.Data.UserID)
		if err != nil {
			log.Errorf(c, err.Error())
			if dal.IsNotFound(err) {
				return nil
			}
			return err
		}

		log.Debugf(c, "counterpartyUser: %v", *counterpartyUser.Data)

		if err := dtdal.DB.RunInTransaction(c, func(tc context.Context) error {
			transfer, err := facade.Transfers.GetTransferByID(tc, tx, transferID)
			if err != nil {
				return err
			}
			changed := false

			// TODO: allow to pass creator counterparty as well. Match by userID

			log.Debugf(c, "transfer.From() before: %v", transfer.Data.From())
			log.Debugf(c, "transfer.To() before: %v", transfer.Data.To())

			// Update transfer creator
			{
				transferCreator := transfer.Data.Creator()
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
				if transferCreator.ContactName == "" || transferCreator.ContactName != counterpartyCounterparty.Data.FullName() {
					transferCreator.ContactName = counterpartyCounterparty.Data.FullName()
					changed = true
				}
				log.Debugf(c, "transferCreator after: %v", transferCreator)
				log.Debugf(c, "transfer.Creator() after: %v", transfer.Data.Creator())
			}

			// Update transfer counterparty
			{
				transferCounterparty := transfer.Data.Counterparty()
				log.Debugf(c, "transferCounterparty before: %v", transferCounterparty)
				if transferCounterparty.UserID == 0 {
					transferCounterparty.UserID = counterpartyCounterparty.Data.UserID
					changed = true
				} else if transferCounterparty.UserID != counterpartyCounterparty.Data.UserID {
					err = fmt.Errorf("transferCounterparty.UserID != counterpartyCounterparty.UserID: %d != %d", transferCounterparty.UserID, counterpartyCounterparty.Data.UserID)
					return err
				} else {
					log.Debugf(c, "transferCounterparty.UserID == counterpartyCounterparty.UserID: %d", transferCounterparty.UserID)
				}
				if transferCounterparty.UserName == "" || transferCounterparty.UserName != counterpartyUser.Data.FullName() {
					transferCounterparty.UserName = counterpartyUser.Data.FullName()
					changed = true
				}
				log.Debugf(c, "transferCounterparty after: %v", transferCounterparty)
				log.Debugf(c, "transfer.Contact() after: %v", transfer.Data.Counterparty())
			}
			log.Debugf(c, "transfer.From() after: %v", transfer.Data.From())
			log.Debugf(c, "transfer.To() after: %v", transfer.Data.To())

			if changed {
				if err = facade.Transfers.SaveTransfer(tc, tx, transfer); err != nil {
					return err
				}
				if !transfer.Data.DtDueOn.IsZero() {
					var counterpartyUser models.AppUser
					if counterpartyUser, err = facade.User.GetUserByID(c, tx, counterpartyCounterparty.Data.UserID); err != nil {
						return err
					}

					if !counterpartyUser.Data.HasDueTransfers {
						if err = dtdal.User.DelayUpdateUserHasDueTransfers(tc, counterpartyCounterparty.Data.UserID); err != nil {
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

	user, err := facade.User.GetUserByID(c, tx, userID)
	if err != nil {
		log.Errorf(c, err.Error())
		if dal.IsNotFound(err) {
			err = nil
		}
		return err
	}

	userName := user.Data.FullName()

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
			err := dtdal.DB.RunInTransaction(c, func(c context.Context) error {
				transfer, err := facade.Transfers.GetTransferByID(c, transferID)
				if err != nil {
					return err
				}
				changed := false
				switch userID {
				case transfer.Data.From().UserID:
					if from := transfer.Data.From(); from.UserName != userName {
						from.UserName = userName
						changed = true
					}
				case transfer.Data.To().UserID:
					if to := transfer.Data.To(); to.UserName != userName {
						to.UserName = userName
						changed = true
					}
				default:
					log.Infof(c, "Transfer(%d) creator is not a counterparty")
				}
				if changed {
					if err = facade.Transfers.SaveTransfer(c, tx, transfer); err != nil {
						return err
					}
				}
				return err
			}, dtdal.SingleGroupTransaction)
			if err != nil {
				log.Errorf(c, err.Error())
			}
			return
		}(key.IntID())
	}
})
