package inspector

import (
	"fmt"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"strconv"

	"sync"

	"time"

	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/strongo/decimal"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type transfersPage struct {
}

func (h transfersPage) transfersPageHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	now := time.Now()

	urlQuery := r.URL.Query()

	currency := money.Currency(urlQuery.Get("currency"))

	contactID, err := strconv.ParseInt(urlQuery.Get("contact"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
	}

	var (
		user                          models.AppUser
		contact                       models.Contact
		transfers                     []models.Transfer
		transfersTotalWithoutInterest decimal.Decimal64p2
	)

	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if contact, err = facade.GetContactByID(c, nil, contactID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, err)
			return
		}
		if user, err = facade.User.GetUserByID(c, nil, contact.Data.UserID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db, err := facade.GetDatabase(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, err)
			return
		}
		if transfers, transfersTotalWithoutInterest, err = h.processTransfers(c, db, contactID, currency); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, err)
			return
		}
	}()

	wg.Wait()

	balancesWithoutInterest := balanceRow{
		user:      user.Data.ContactByID(contactID).Balance()[currency],
		contacts:  contact.Data.Balance()[currency],
		transfers: transfersTotalWithoutInterest,
	}

	balancesWithInterest := balanceRow{}
	if balance, err := user.Data.ContactByID(contactID).BalanceWithInterest(c, now); err == nil {
		balancesWithInterest.user = balance[currency]
	} else {
		balancesWithInterest.userContactBalanceErr = err
	}
	if balance, err := contact.Data.BalanceWithInterest(c, now); err == nil {
		balancesWithInterest.contacts = balance[currency]
	} else {
		balancesWithInterest.contactBalanceErr = err
	}

	renderTransfersPage(contact, currency, balancesWithoutInterest, balancesWithInterest, transfers, w)
}

func (h transfersPage) processTransfers(c context.Context, tx dal.ReadSession, contactID int64, currency money.Currency) (
	transfers []models.Transfer,
	balanceWithoutInterest decimal.Decimal64p2,
	err error,
) {
	query := dal.From(models.TransferKind).
		Where(
			dal.Field("BothCounterpartyIDs").EqualTo(contactID),
			dal.Field("Currency").EqualTo(currency),
		).
		OrderBy(dal.DescendingField("DtCreated")).
		SelectInto(models.NewTransferRecord)

	var reader dal.Reader
	if reader, err = tx.QueryReader(c, query); err != nil {
		return
	}
	for {
		var record dal.Record
		if record, err = reader.Next(); err != nil {
			if err == datastore.Done {
				err = nil
				break
			}
			panic(err)
		}
		transfer := models.NewTransfer(record.Key().ID.(int), record.Data().(*models.TransferData))
		transfers = append(transfers, transfer)
		switch contactID {
		case transfer.Data.From().ContactID:
			balanceWithoutInterest -= transfer.Data.AmountInCents
		case transfer.Data.To().ContactID:
			balanceWithoutInterest += transfer.Data.AmountInCents
		default:
			panic(fmt.Sprintf("contactID != from && contactID != to: contactID=%v, from=%v, to=%v",
				contactID, transfer.Data.From().ContactID, transfer.Data.To().ContactID))
		}
	}

	return
}
