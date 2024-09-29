package unsorted

import (
	"context"
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-core-modules/auth/token4auth"
	"github.com/sneat-co/sneat-core-modules/auth/unsorted4auth"
	common4all2 "github.com/sneat-co/sneat-core-modules/common4all"
	dal4contactus2 "github.com/sneat-co/sneat-core-modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-core-modules/core/queues"
	"github.com/sneat-co/sneat-core-modules/userus/dal4userus"
	"github.com/sneat-co/sneat-core-modules/userus/dbo4userus"
	"github.com/sneat-co/sneat-go-core/facade"
	facade4debtus2 "github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus/dto4debtus"
	models4debtus2 "github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/strongo/delaying"
	"github.com/strongo/logus"
	"github.com/strongo/validation"
	"net/http"
	"reflect"
)

func HandleAdminFindUser(ctx context.Context, w http.ResponseWriter, r *http.Request, _ token4auth.AuthInfo) {

	if userID := r.URL.Query().Get("userID"); userID != "" {
		appUser := dbo4userus.NewUserEntry(userID)
		if err := dal4userus.GetUser(ctx, nil, appUser); err != nil {
			logus.Errorf(ctx, fmt.Errorf("failed to get user by userID=%s: %w", userID, err).Error())
		} else {
			common4all2.JsonToResponse(ctx, w, []dto4debtus.ApiUserDto{{ID: userID, Name: appUser.Data.GetFullName()}})
		}
		return
	} else {
		tgUserText := r.URL.Query().Get("tgUser")

		if tgUserText == "" {
			common4all2.BadRequestMessage(ctx, w, "tgUser is empty string")
			return
		}

		tgUsers, err := unsorted4auth.TgUser.FindByUserName(ctx, nil, tgUserText)

		if err != nil {
			common4all2.InternalError(ctx, w, err)
			return
		}

		users := make([]dto4debtus.ApiUserDto, len(tgUsers))

		for i, tgUser := range tgUsers {
			users[i] = dto4debtus.ApiUserDto{
				ID:   tgUser.GetAppUserID(),
				Name: tgUser.BaseData().UserName,
			}
		}

		common4all2.JsonToResponse(ctx, w, users)
	}
}

func HandleAdminMergeUserContacts(ctx context.Context, w http.ResponseWriter, r *http.Request, _ token4auth.AuthInfo) {
	keepID := common4all2.GetStrID(ctx, w, r, "keepID")
	if keepID == "" {
		return
	}
	deleteID := common4all2.GetStrID(ctx, w, r, "deleteID")
	if deleteID == "" {
		return
	}
	spaceID := common4all2.GetStrID(ctx, w, r, "spaceID")
	if spaceID == "" {
		common4all2.BadRequestError(ctx, w, validation.NewErrRequestIsMissingRequiredField("spaceID"))
		return
	}

	logus.Infof(ctx, "keepID: %s, deleteID: %s", keepID, deleteID)

	if err := facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {

		contacts, err := dal4contactus2.GetContactsByIDs(ctx, tx, spaceID, []string{keepID, deleteID})
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
		if contactToDelete.Data.UserID != "" && contactToKeep.Data.UserID == "" {
			return errors.New("contactToDelete.CounterpartyUserID != 0 && contactToKeep.CounterpartyUserID == 0")
		}
		contactusSpace := dal4contactus2.NewContactusSpaceEntry(spaceID)

		if err = dal4contactus2.GetContactusSpace(ctx, tx, contactusSpace); err != nil {
			return err
		}

		if contactusSpace.Data.HasContact(deleteID) {
			update := contactusSpace.Data.RemoveContact(deleteID)
			if err = tx.Update(ctx, contactusSpace.Key, []dal.Update{update}); err != nil {
				return err
			}
		}
		if err := delayChangeTransfersCounterparty.EnqueueWork(ctx, delaying.With(queues.QueueSupport, "changeTransfersCounterparty", 0), deleteID, keepID, ""); err != nil {
			return err
		}
		if err := tx.Delete(ctx, models4debtus2.NewDebtusContactKey(spaceID, deleteID)); err != nil {
			return err
		} else {
			logus.Warningf(ctx, "DebtusSpaceContactEntry %s has been deleted from DB (non revocable)", deleteID)
		}
		return nil
	}); err != nil {
		common4all2.ErrorAsJson(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

func DelayedChangeTransfersCounterparty(ctx context.Context, oldID, newID int64, cursor string) (err error) {
	logus.Debugf(ctx, "delayedChangeTransfersCounterparty(oldID=%d, newID=%d)", oldID, newID)

	var q = dal.From(models4debtus2.TransfersCollection).
		WhereField("BothCounterpartyIDs", dal.Equal, oldID).
		Limit(100).
		SelectKeysOnly(reflect.Int)

	var db dal.DB

	if db, err = facade.GetSneatDB(ctx); err != nil {
		return err
	}

	var reader dal.Reader
	if reader, err = db.QueryReader(ctx, q); err != nil {
		return err
	}
	transferIDs, err := dal.SelectAllIDs[int](reader, dal.WithLimit(q.Limit()))
	if err != nil {
		return err
	}

	logus.Infof(ctx, "Loaded %d transferIDs", len(transferIDs))
	args := make([][]interface{}, len(transferIDs))
	for i, id := range transferIDs {
		args[i] = []interface{}{id, oldID, newID, ""}
	}
	return delayChangeTransferCounterparty.EnqueueWorkMulti(ctx, delaying.With(queues.QueueSupport, "changeTransferCounterparty", 0), args...)
}

func DelayedChangeTransferCounterparty(ctx context.Context, spaceID, transferID, oldID, newID string, cursor string) (err error) {
	logus.Debugf(ctx, "delayedChangeTransferCounterparty(spaceID=%s, oldID=%s, newID=%s, cursor=%s)", spaceID, oldID, newID, cursor)
	if _, err = facade4debtus2.GetDebtusSpaceContactByID(ctx, nil, spaceID, newID); err != nil {
		return err
	}
	err = facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		transfer, err := facade4debtus2.Transfers.GetTransferByID(ctx, tx, transferID)
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
			err = facade4debtus2.Transfers.SaveTransfer(ctx, tx, transfer)
		}
		return err
	})
	return err
}
