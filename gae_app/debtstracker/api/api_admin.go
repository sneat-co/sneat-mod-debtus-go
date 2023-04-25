package api

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	apphostgae "github.com/strongo/app-host-gae"
	"net/http"
	"reflect"
	"strconv"

	"context"
	"errors"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/api/dto"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/auth"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
)

func handleAdminFindUser(c context.Context, w http.ResponseWriter, r *http.Request, _ auth.AuthInfo) {

	if userID := r.URL.Query().Get("userID"); userID != "" {
		if user, err := dtdal.User.GetUserByStrID(c, userID); err != nil {
			log.Errorf(c, fmt.Errorf("failed to get user by ID=%v: %w", userID, err).Error())
		} else {
			jsonToResponse(c, w, []dto.ApiUserDto{{ID: userID, Name: user.Data.FullName()}})
		}
		return
	} else {
		tgUserText := r.URL.Query().Get("tgUser")

		if tgUserText == "" {
			BadRequestMessage(c, w, "tgUser is empty string")
			return
		}

		tgUsers, err := dtdal.TgUser.FindByUserName(c, nil, tgUserText)

		if err != nil {
			InternalError(c, w, err)
			return
		}

		users := make([]dto.ApiUserDto, len(tgUsers))

		for i, tgUser := range tgUsers {
			users[i] = dto.ApiUserDto{
				ID:   strconv.FormatInt(tgUser.Data.AppUserIntID, 10),
				Name: tgUser.Data.Name(),
			}
		}

		jsonToResponse(c, w, users)
	}
}

func handleAdminMergeUserContacts(c context.Context, w http.ResponseWriter, r *http.Request, _ auth.AuthInfo) {
	keepID := int64(getID(c, w, r, "keepID"))
	deleteID := int64(getID(c, w, r, "deleteID"))

	log.Infof(c, "keepID: %d, deleteID: %d", keepID, deleteID)

	db, err := facade.GetDatabase(c)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	if err := db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		contacts, err := facade.GetContactsByIDs(c, tx, []int64{keepID, deleteID})
		if err != nil {
			return err
		}
		if len(contacts) < 2 {
			return fmt.Errorf("len(contacts):%d < 2", len(contacts))
		}
		contactToKeep := contacts[0]
		contactToDelete := contacts[1]
		if contactToKeep.Data.UserID != contactToDelete.Data.UserID {
			return errors.New("contactToKeep.UserID != contactToDelete.UserID")
		}
		if contactToDelete.Data.CounterpartyUserID != 0 && contactToKeep.Data.CounterpartyUserID == 0 {
			return errors.New("contactToDelete.CounterpartyUserID != 0 && contactToKeep.CounterpartyUserID == 0")
		}
		user, err := facade.User.GetUserByID(c, tx, contactToKeep.Data.UserID)
		if err != nil {
			return err
		}
		if user.ID != 0 {
			return errors.New("Not implemented yet: Need to update counterparty & user balances + last transfer info")
		}
		if userChanged := user.Data.RemoveContact(deleteID); userChanged {
			if err = facade.User.SaveUser(c, tx, user); err != nil {
				return err
			}
		}
		if err := apphostgae.EnqueueWork(c, common.QUEUE_SUPPORT, "changeTransfersCounterparty", 0, delayedChangeTransfersCounterparty, deleteID, keepID, ""); err != nil {
			return err
		}
		if err := tx.Delete(c, models.NewContactKey(deleteID)); err != nil {
			return err
		} else {
			log.Warningf(c, "Contact %d has been deleted from DB (non revocable)", deleteID)
		}
		return nil
	}); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}

var delayedChangeTransfersCounterparty = delay.Func("changeTransfersCounterparty", func(c context.Context, oldID, newID int64, cursor string) (err error) {
	log.Debugf(c, "delayedChangeTransfersCounterparty(oldID=%d, newID=%d)", oldID, newID)

	var q = dal.From(models.TransferKind).
		WhereField("BothCounterpartyIDs", dal.Equal, oldID).
		Limit(100).
		SelectKeysOnly(reflect.Int)

	var reader dal.Reader
	if reader, err = facade.DB().QueryReader(c, q); err != nil {
		return err
	}
	transferIDs, err := dal.SelectAllIDs[int](reader, q.Limit())
	if err != nil {
		return err
	}

	log.Infof(c, "Loaded %d transferIDs", len(transferIDs))
	args := make([][]interface{}, len(transferIDs))
	for i, id := range transferIDs {
		args[i] = []interface{}{id, oldID, newID, ""}
	}
	return apphostgae.EnqueueWorkMulti(c, common.QUEUE_SUPPORT, "changeTransferCounterparty", 0, delayedChangeTransferCounterparty, args...)
})

var delayedChangeTransferCounterparty = delay.Func("changeTransferCounterparty", func(c context.Context, transferID int, oldID, newID int64, cursor string) (err error) {
	log.Debugf(c, "delayedChangeTransferCounterparty(oldID=%d, newID=%d, cursor=%v)", oldID, newID, cursor)
	if _, err = facade.GetContactByID(c, nil, newID); err != nil {
		return err
	}
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return err
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		transfer, err := facade.Transfers.GetTransferByID(c, tx, transferID)
		if err != nil {
			return err
		}
		changed := false
		for i, contactID := range transfer.Data.BothCounterpartyIDs {
			if contactID == oldID {
				transfer.Data.BothCounterpartyIDs[i] = newID
				changed = true
				break
			}
		}
		if changed {
			if from := transfer.Data.From(); from.ContactID == oldID {
				from.ContactID = newID
			} else if to := transfer.Data.To(); to.ContactID == oldID {
				to.ContactID = newID
			}
			err = facade.Transfers.SaveTransfer(c, tx, transfer)
		}
		return err
	})
	return err
})
