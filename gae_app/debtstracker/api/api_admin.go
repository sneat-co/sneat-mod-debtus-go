package api

import (
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/api/dto"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal/gaedal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/log"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
)

func handleAdminFindUser(c context.Context, w http.ResponseWriter, r *http.Request, _ auth.AuthInfo) {

	if userID := r.URL.Query().Get("userID"); userID != "" {
		if user, err := dal.User.GetUserByStrID(c, userID); err != nil {
			log.Errorf(c, errors.Wrapf(err, "Failed to get user by ID=%v", userID).Error())
		} else {
			jsonToResponse(c, w, []dto.ApiUserDto{{ID: userID, Name: user.FullName()}})
		}
		return
	} else {
		tgUserText := r.URL.Query().Get("tgUser")

		if tgUserText == "" {
			BadRequestMessage(c, w, "tgUser is empty string")
			return
		}

		tgUsers, err := dal.TgUser.FindByUserName(c, tgUserText)

		if err != nil {
			InternalError(c, w, err)
			return
		}

		users := make([]dto.ApiUserDto, len(tgUsers))

		for i, tgUser := range tgUsers {
			users[i] = dto.ApiUserDto{
				ID:   strconv.FormatInt(tgUser.AppUserIntID, 10),
				Name: tgUser.Name(),
			}
		}

		jsonToResponse(c, w, users)
	}
}

func handleAdminMergeUserContacts(c context.Context, w http.ResponseWriter, r *http.Request, _ auth.AuthInfo) {
	keepID := getID(c, w, r, "keepID")
	deleteID := getID(c, w, r, "deleteID")

	log.Infof(c, "keepID: %d, deleteID: %d", keepID, deleteID)

	if err := dal.DB.RunInTransaction(c, func(c context.Context) error {
		contacts, err := facade.GetContactsByIDs(c, []int64{keepID, deleteID})
		if err != nil {
			return err
		}
		if len(contacts) < 2 {
			return fmt.Errorf("len(contacts):%d < 2", len(contacts))
		}
		contactToKeep := contacts[0]
		contactToDelete := contacts[1]
		if contactToKeep.UserID != contactToDelete.UserID {
			return errors.New("contactToKeep.UserID != contactToDelete.UserID")
		}
		if contactToDelete.CounterpartyUserID != 0 && contactToKeep.CounterpartyUserID == 0 {
			return errors.New("contactToDelete.CounterpartyUserID != 0 && contactToKeep.CounterpartyUserID == 0")
		}
		user, err := facade.User.GetUserByID(c, contactToKeep.UserID)
		if err != nil {
			return err
		}
		if user.ID != 0 {
			return errors.New("Not implemented yet: Need to update counterparty & user balances + last transfer info")
		}
		if userChanged := user.RemoveContact(deleteID); userChanged {
			if err = facade.User.SaveUser(c, user); err != nil {
				return err
			}
		}
		if task, err := gae.CreateDelayTask(common.QUEUE_SUPPORT, "changeTransfersCounterparty", delayedChangeTransfersCounterparty, deleteID, keepID, ""); err != nil {
			return err
		} else {
			if _, err = taskqueue.Add(c, task, common.QUEUE_SUPPORT); err != nil {
				return err
			}
		}
		if err := datastore.Delete(c, gaedal.NewContactKey(c, deleteID)); err != nil {
			return err
		} else {
			log.Warningf(c, "Contact %d has been deleted from DB (non revocable)", deleteID)
		}
		return nil
	}, dal.CrossGroupTransaction); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}

var delayedChangeTransfersCounterparty = delay.Func("changeTransfersCounterparty", func(c context.Context, oldID, newID int64, cursor string) error {
	log.Debugf(c, "delayedChangeTransfersCounterparty(oldID=%d, newID=%d)", oldID, newID)

	query := datastore.NewQuery(models.TransferKind).Filter("BothCounterpartyIDs =", oldID).Limit(100)
	query = query.KeysOnly()
	if keys, err := query.GetAll(c, nil); err != nil {
		return err
	} else {
		log.Infof(c, "Loaded %d keys", len(keys))
		tasks := make([]*taskqueue.Task, len(keys))
		for i, key := range keys {
			if tasks[i], err = gae.CreateDelayTask(common.QUEUE_SUPPORT, "changeTransferCounterparty", delayedChangeTransferCounterparty, key.IntID(), oldID, newID, ""); err != nil {
				return err
			}
		}
		if _, err = taskqueue.AddMulti(c, tasks, common.QUEUE_SUPPORT); err != nil {
			return err
		}
	}
	return nil
})

var delayedChangeTransferCounterparty = delay.Func("changeTransferCounterparty", func(c context.Context, transferID, oldID, newID int64, cursor string) error {
	log.Debugf(c, "delayedChangeTransferCounterparty(oldID=%d, newID=%d, cursor=%v)", oldID, newID, cursor)
	if _, err := facade.GetContactByID(c, newID); err != nil {
		return err
	}
	err := dal.DB.RunInTransaction(c, func(c context.Context) error {
		transfer, err := facade.GetTransferByID(c, transferID)
		if err != nil {
			return err
		}
		changed := false
		for i, contactID := range transfer.BothCounterpartyIDs {
			if contactID == oldID {
				transfer.BothCounterpartyIDs[i] = newID
				changed = true
				break
			}
		}
		if changed {
			if from := transfer.From(); from.ContactID == oldID {
				from.ContactID = newID
			} else if to := transfer.To(); to.ContactID == oldID {
				to.ContactID = newID
			}
			err = facade.Transfers.SaveTransfer(c, transfer)
		}
		return err
	}, dal.SingleGroupTransaction)
	return err
})
