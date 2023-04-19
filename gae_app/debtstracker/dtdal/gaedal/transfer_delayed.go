package gaedal

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"reflect"
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
	"google.golang.org/appengine/delay"
)

func (TransferDalGae) DelayUpdateTransfersWithCounterparty(c context.Context, creatorCounterpartyID, counterpartyCounterpartyID int64) (err error) {
	log.Debugf(c, "DelayUpdateTransfersWithCounterparty(creatorCounterpartyID=%d, counterpartyCounterpartyID=%d)", creatorCounterpartyID, counterpartyCounterpartyID)
	if creatorCounterpartyID == 0 {
		return errors.New("creatorCounterpartyID == 0")
	}
	if counterpartyCounterpartyID == 0 {
		return errors.New("counterpartyCounterpartyID == 0")
	}
	if task, err := gae.CreateDelayTask(common.QUEUE_TRANSFERS, DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY, delayedUpdateTransfersWithCounterparty, creatorCounterpartyID, counterpartyCounterpartyID); err != nil {
		return err
	} else if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_TRANSFERS); err != nil {
		return err
	}
	return nil
}

const (
	DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY  = "update-transfers-with-counterparty"
	DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY = "update-1-transfer-with-counterparty"
)

var delayedUpdateTransfersWithCounterparty = delay.Func(DELAY_UPDATE_TRANSFERS_WITH_COUNTERPARTY, func(c context.Context, creatorCounterpartyID, counterpartyCounterpartyID int64) (err error) {
	log.Infof(c, "delayedUpdateTransfersWithCounterparty(creatorCounterpartyID=%d, counterpartyCounterpartyID=%d)", creatorCounterpartyID, counterpartyCounterpartyID)
	if creatorCounterpartyID == 0 {
		log.Errorf(c, "creatorCounterpartyID == 0")
		return nil
	}
	if counterpartyCounterpartyID == 0 {
		log.Errorf(c, "counterpartyCounterpartyID == 0")
		return nil
	}

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	query := dal.From(models.TransferKind).
		WhereField("BothCounterpartyIDs", dal.Equal, creatorCounterpartyID).WhereField("BothCounterpartyIDs", dal.Equal, 0).
		OrderBy(dal.DescendingField("DtCreated")).
		SelectKeysOnly(reflect.Int)

	var reader dal.Reader
	if reader, err = db.QueryReader(c, query); err != nil {
		return err
	}
	if transferIDs, err := dal.SelectAllIDs[int](reader, query.Limit); err != nil {
		return fmt.Errorf("failed to load transfers: %w", err)
	} else if len(transferIDs) > 0 {
		log.Infof(c, "Loaded %d transfer IDs", len(transferIDs))
		delayDuration := 10 * time.Microsecond
		for _, transferID := range transferIDs {
			if task, err := gae.CreateDelayTask(common.QUEUE_TRANSFERS, DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY, delayedUpdateTransferWithCounterparty, transferID, counterpartyCounterpartyID); err != nil {
				return fmt.Errorf("failed to create task for transfer id=%d: %w", transferID, err)
			} else {
				task.Delay = delayDuration
				delayDuration += 10 * time.Microsecond
				if _, err = gae.AddTaskToQueue(c, task, common.QUEUE_TRANSFERS); err != nil {
					return fmt.Errorf("failed to add task for transfer %d to queue [%v]: %w", transferID, common.QUEUE_TRANSFERS, err)
				}
			}
		}
	} else {
		query := dal.From(models.TransferKind).
			WhereField("BothCounterpartyIDs", dal.Equal, creatorCounterpartyID).WhereField("BothCounterpartyIDs", dal.Equal, counterpartyCounterpartyID).
			SelectKeysOnly(reflect.Int)
		query.Limit = 1
		var reader dal.Reader
		if reader, err = db.QueryReader(c, query); err != nil {
			return err
		}
		var transferIDs []int
		if transferIDs, err = dal.SelectAllIDs[int](reader, query.Limit); err != nil {
			return fmt.Errorf("failed to load transfers by 2 counterparty IDs: %w", err)
		}
		if len(transferIDs) > 0 {
			log.Infof(c, "No transfers found to update counterparty details")
		} else {
			log.Warningf(c, "No transfers found to update counterparty details")
		}
	}
	return nil
})

var delayedUpdateTransferWithCounterparty = delay.Func(DELAY_UPDATE_1_TRANSFER_WITH_COUNTERPARTY,
	func(c context.Context, transferID int, counterpartyCounterpartyID int64) (err error) {
		log.Debugf(c, "delayedUpdateTransferWithCounterparty(transferID=%d, counterpartyCounterpartyID=%d)", transferID, counterpartyCounterpartyID)
		if transferID == 0 {
			log.Errorf(c, "transferID == 0")
			return nil
		}
		if counterpartyCounterpartyID == 0 {
			log.Errorf(c, "counterpartyCounterpartyID == 0")
			return nil
		}

		var db dal.Database
		if db, err = GetDatabase(c); err != nil {
			return err
		}

		counterpartyCounterparty, err := facade.GetContactByID(c, db, counterpartyCounterpartyID)
		if err != nil {
			log.Errorf(c, err.Error())
			if dal.IsNotFound(err) {
				return nil
			}
			return err
		}

		log.Debugf(c, "counterpartyCounterparty: %v", counterpartyCounterparty)

		counterpartyUser, err := facade.User.GetUserByID(c, db, counterpartyCounterparty.Data.UserID)
		if err != nil {
			log.Errorf(c, err.Error())
			if dal.IsNotFound(err) {
				return nil
			}
			return err
		}

		log.Debugf(c, "counterpartyUser: %v", *counterpartyUser.Data)

		if err := db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) error {
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

var delayedUpdateTransfersWithCreatorName = delay.Func(UPDATE_TRANSFERS_WITH_CREATOR_NAME, func(c context.Context, userID int64) (err error) {
	log.Debugf(c, "delayedUpdateTransfersWithCreatorName(userID=%d)", userID)

	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return err
	}
	user, err := facade.User.GetUserByID(c, db, userID)
	if err != nil {
		log.Errorf(c, err.Error())
		if dal.IsNotFound(err) {
			err = nil
		}
		return err
	}

	userName := user.Data.FullName()

	query := dal.From(models.TransferKind).
		WhereField("BothUserIDs", dal.Equal, userID).
		SelectInto(models.NewTransferRecord)

	var reader dal.Reader
	reader, err = db.QueryReader(c, query)

	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		transferRecord, err := reader.Next()
		if err != nil {
			return err
		}
		trasfer := models.TransferFromRecord(transferRecord)
		if err != nil {
			if err == dal.ErrNoMoreRecords {
				return nil
			}
			log.Errorf(c, err.Error())
			return err
		}
		wg.Add(1)
		go func(transferID int) {
			defer wg.Done()
			err := db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
				transfer, err := facade.Transfers.GetTransferByID(c, tx, transferID)
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
			})
			if err != nil {
				log.Errorf(c, err.Error())
			}
		}(trasfer.ID)
	}
})
