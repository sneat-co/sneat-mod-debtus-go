package inspector

import (
	"fmt"
	"net/http"
	"strconv"

	"sync"

	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/decimal"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type transfersPage struct {
}

func (h transfersPage) transfersPageHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	now := time.Now()

	urlQuery := r.URL.Query()

	currency := models.Currency(urlQuery.Get("currency"))

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
		if contact, err = dal.Contact.GetContactByID(c, contactID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		if user, err = dal.User.GetUserByID(c, contact.UserID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if transfers, transfersTotalWithoutInterest, err = h.processTransfers(c, contactID, currency); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
	}()

	wg.Wait()

	balancesWithoutInterest := balanceRow{
		user:      user.ContactByID(contactID).Balance()[currency],
		contacts:  contact.Balance()[currency],
		transfers: transfersTotalWithoutInterest,
	}

	balancesWithInterest := balanceRow{}
	if balance, err := user.ContactByID(contactID).BalanceWithInterest(c, now); err == nil {
		balancesWithInterest.user = balance[currency]
	} else {
		balancesWithInterest.userContactBalanceErr = err
	}
	if balance, err := contact.BalanceWithInterest(c, now); err == nil {
		balancesWithInterest.contacts = balance[currency]
	} else {
		balancesWithInterest.contactBalanceErr = err
	}

	renderTransfersPage(contact, currency, balancesWithoutInterest, balancesWithInterest, transfers, w)
}

func (h transfersPage) processTransfers(c context.Context, contactID int64, currency models.Currency) (
	transfers []models.Transfer,
	balanceWithoutInterest decimal.Decimal64p2,
	err error,
) {
	query := datastore.NewQuery(models.TransferKind)
	query = query.Filter("BothCounterpartyIDs=", contactID)
	query = query.Filter("Currency=", currency)
	query = query.Order("DtCreated")

	iterator := query.Run(c)

	for {
		transferEntity := new(models.TransferEntity)
		var key *datastore.Key
		if key, err = iterator.Next(transferEntity); err != nil {
			if err == datastore.Done {
				err = nil
				break
			}
			panic(err)
		}
		transfer := models.NewTransfer(key.IntID(), transferEntity)
		transfers = append(transfers, transfer)
		switch contactID {
		case transfer.From().ContactID:
			balanceWithoutInterest -= transfer.AmountInCents
		case transfer.To().ContactID:
			balanceWithoutInterest += transfer.AmountInCents
		default:
			panic(fmt.Sprintf("contactID != from && contactID != to: contactID=%v, from=%v, to=%v",
				contactID, transfer.From().ContactID, transfer.To().ContactID))
		}
	}

	return
}
