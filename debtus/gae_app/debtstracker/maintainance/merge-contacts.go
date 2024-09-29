package maintainance

import (
	"context"
	"errors"
	"fmt"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	dal4contactus2 "github.com/sneat-co/sneat-core-modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-go-core/facade"
	facade4debtus2 "github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus"
	models4debtus2 "github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/strongo/logus"
	"net/http"
	"strings"
	"sync"
)

func mergeContactsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	targetContactID := q.Get("targetContactID")
	if targetContactID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("target contact ContactID is empty"))
		return
	}
	spaceID := q.Get("spaceID")
	sourceContactIDs := strings.Split(q.Get("sourceContactIDs"), ",")
	var db dal.DB
	var err error
	db, err = facade.GetSneatDB(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	ctx := r.Context()
	userCtx := facade.NewUserContext("")
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		return mergeContacts(ctx, userCtx, tx, spaceID, targetContactID, sourceContactIDs...)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write([]byte("done"))
}

func mergeContacts(ctx context.Context, userCtx facade.UserContext, tx dal.ReadwriteTransaction, spaceID, targetContactID string, sourceContactIDs ...string) (err error) {
	if len(sourceContactIDs) == 0 {
		panic("len(sourceContactIDs) == 0")
	}

	contactusSpace := dal4contactus2.NewContactusSpaceEntry(spaceID)
	debtusSpace := models4debtus2.NewDebtusSpaceEntry(spaceID)

	targetContact := dal4contactus2.NewContactEntry(spaceID, targetContactID)

	var targetDebtusContact models4debtus2.DebtusSpaceContactEntry

	if targetDebtusContact, err = facade4debtus2.GetDebtusSpaceContactByID(ctx, tx, spaceID, targetContactID); err != nil {
		if dal.IsNotFound(err) && len(sourceContactIDs) == 1 {
			if targetDebtusContact, err = facade4debtus2.GetDebtusSpaceContactByID(ctx, tx, spaceID, sourceContactIDs[0]); err != nil {
				return
			}
			targetDebtusContact.ID = targetContactID
			if err = facade4debtus2.SaveContact(ctx, targetDebtusContact); err != nil {
				return
			}
		} else {
			return
		}
	}

	if err = dal4contactus2.GetContactusSpace(ctx, tx, contactusSpace); err != nil {
		return
	}

	for _, sourceContactID := range sourceContactIDs {
		if sourceContactID == targetContactID {
			err = fmt.Errorf("sourceContactID == targetContactID: %v", sourceContactID)
			return
		}
		sourceContact := dal4contactus2.NewContactEntry(spaceID, sourceContactID)
		if err = dal4contactus2.GetContact(ctx, tx, sourceContact); err != nil {
			if dal.IsNotFound(err) {
				continue
			}
			return
		}
		if sourceContact.Data.UserID != targetContact.Data.UserID {
			err = fmt.Errorf("sourceDebtusContact.UserID != targetDebtusContact.UserID: %v != %v",
				sourceContact.Data.UserID, targetContact.Data.UserID)
			return
		}
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(sourceContactIDs))

	for _, sourceContactID := range sourceContactIDs {
		go func(sourceContactID string) {
			if err2 := mergeContactTransfers(ctx, tx, wg, targetContactID, sourceContactID); err2 != nil {
				logus.Errorf(ctx, "failed to merge api4transfers for contact %v: %v", sourceContactID, err2)
				if err == nil {
					err = err2
				}
			}
		}(sourceContactID)
	}
	wg.Wait()

	if err != nil {
		return
	}

	if err = facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		if err = dal4contactus2.GetContactusSpace(ctx, tx, contactusSpace); err != nil {
			return
		}
		debtusContacts := make(map[string]*models4debtus2.DebtusContactBrief)
		targetContactBalance := targetDebtusContact.Data.Balance
		for contactID, debtusContact := range debtusSpace.Data.Contacts {
			for _, sourceContactID := range sourceContactIDs {
				if contactID == sourceContactID {
					for currency, value := range debtusContact.Balance {
						targetContactBalance.Add(money.NewAmount(currency, value))
					}
					sourceDebtusContact := models4debtus2.NewDebtusSpaceContactEntry(spaceID, sourceContactID, nil)
					if err = facade4debtus2.GetDebtusSpaceContact(ctx, tx, sourceDebtusContact); err != nil {
						if !dal.IsNotFound(err) {
							return
						}
					} else {
						targetDebtusContact.Data.CountOfTransfers += sourceDebtusContact.Data.CountOfTransfers
						if targetDebtusContact.Data.LastTransferAt.Before(sourceDebtusContact.Data.LastTransferAt) {
							targetDebtusContact.Data.LastTransferAt = sourceDebtusContact.Data.LastTransferAt
							targetDebtusContact.Data.LastTransferID = sourceDebtusContact.Data.LastTransferID
						}
						if sourceDebtusContact.Data.CounterpartyContactID != "" {
							counterpartyContact := models4debtus2.NewDebtusSpaceContactEntry(spaceID, sourceDebtusContact.Data.CounterpartyContactID, nil)
							if err = facade4debtus2.GetDebtusSpaceContact(ctx, tx, counterpartyContact); err != nil {
								if !dal.IsNotFound(err) {
									return
								}
							} else if counterpartyContact.Data.CounterpartyContactID == sourceContactID {
								counterpartyContact.Data.CounterpartyContactID = targetContactID
								if err = facade4debtus2.SaveContact(ctx, counterpartyContact); err != nil {
									return
								}
							} else if counterpartyContact.Data.CounterpartyContactID != "" && counterpartyContact.Data.CounterpartyContactID != targetContactID {
								err = fmt.Errorf(
									"data integrity issue : counterpartyContact(id=%v).CounterpartyContactID != sourceContactID: %v != %v",
									counterpartyContact.ID, counterpartyContact.Data.CounterpartyContactID, sourceContactID)
								return
							}
						}
					}
					if err = facade4debtus2.DeleteContactTx(ctx, userCtx, tx, spaceID, sourceContactID); err != nil {
						return
					}
				} else {
					debtusContacts[contactID] = debtusContact
				}
			}
		}
		if debtusTargetContact := debtusContacts[targetContactID]; debtusTargetContact != nil {
			debtusTargetContact.Balance = targetContactBalance
			debtusSpace.Data.SetContacts(debtusContacts)
		}

		if err = tx.SetMulti(ctx, []dal.Record{debtusSpace.Record, contactusSpace.Record}); err != nil {
			return
		}
		return
	}); err != nil {
		return fmt.Errorf("%w: failed to update contactusSpace entity", err)
	}

	return
}

func mergeContactTransfers(ctx context.Context, tx dal.ReadwriteTransaction, wg *sync.WaitGroup, targetContactID string, sourceContactID string) (err error) {
	defer func() {
		wg.Done()
	}()
	transfersQ := dal.From(models4debtus2.TransfersCollection).
		Where(dal.Field("BothCounterpartyIDs").EqualTo(sourceContactID)).
		SelectInto(func() dal.Record {
			return models4debtus2.NewTransfer("", nil).Record
		})
	transfers, err := tx.QueryReader(ctx, transfersQ)
	if err != nil {
		return fmt.Errorf("failed to select api4transfers: %w", err)
	}
	var (
		record   dal.Record
		transfer models4debtus2.TransferEntry
	)
	for {
		if record, err = transfers.Next(); err != nil {
			if errors.Is(err, dal.ErrNoMoreRecords) {
				err = nil
				break
			}
			logus.Errorf(ctx, "Failed to get next transfer: %v", err)
		}
		transfer.ID = record.Key().ID.(string)
		switch sourceContactID {
		case transfer.Data.From().ContactID:
			transfer.Data.From().ContactID = targetContactID
		case transfer.Data.To().ContactID:
			transfer.Data.To().ContactID = targetContactID
		}
		switch sourceContactID {
		case transfer.Data.BothCounterpartyIDs[0]:
			transfer.Data.BothCounterpartyIDs[0] = targetContactID
		case transfer.Data.BothCounterpartyIDs[1]:
			transfer.Data.BothCounterpartyIDs[1] = targetContactID
		}
		if err = facade4debtus2.Transfers.SaveTransfer(ctx, tx, transfer); err != nil {
			logus.Errorf(ctx, "Failed to save transfer #%v: %v", transfer.ID, err)
		}
	}
	return
}
