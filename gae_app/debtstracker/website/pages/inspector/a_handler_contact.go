package inspector

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"strconv"
	//"sync"

	"sync"

	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type contactPage struct {
}

func (h contactPage) contactPageHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	contactID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, err)
		return
	}

	var contact models.Contact

	if contact, err = facade.GetContactByID(c, nil, contactID); err != nil {
		_, _ = fmt.Fprint(w, err)
		return
	}

	//var user, counterpartyUser models.AppUser
	//var counterpartyContact models.Contact
	//

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		//var transfers []models.Transfer
		if _, err = h.verifyTransfers(c, contactID); err != nil {
			panic(err)
		}
	}()

	//
	//wg.Add(1)
	//go func() {
	//	if user, err = facade.User.GetUserByID(c, contact.UserID); err != nil {
	//		return
	//	}
	//
	//}()
	//
	//if contact.CounterpartyUserID != 0 {
	//	wg.Add(1)
	//	if user, err = facade.User.GetUserByID(c, contact.CounterpartyUserID); err != nil {
	//		return
	//	}
	//}
	//
	//if contact.CounterpartyCounterpartyID != 0 {
	//	wg.Add(1)
	//	if counterpartyContact, err = facade.GetContactByID(c, tx, contact.CounterpartyCounterpartyID); err != nil {
	//		return
	//	}
	//}

	RenderContactPage(contact, w)

	//renderContactUsers(w, user, counterpartyUser)

}

func (contactPage) verifyTransfers(c context.Context, contactID int64) (
	transfers []models.Transfer, err error,
) {

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	//select := dal.Select{
	//	From: &dal.CollectionRef{Name: models.TransferKind},
	//}
	query := dal.From(models.TransferKind).
		Where(dal.Field("BothCounterpartyIDs").EqualTo(contactID)).
		SelectInto(models.NewTransferRecord)

	var reader dal.Reader
	if reader, err = db.QueryReader(c, query); err != nil {
		return
	}

	for {
		//transferEntity := new(models.TransferData)
		//var key *datastore.Key
		var record dal.Record
		if record, err = reader.Next(); err != nil {
			if err == datastore.Done {
				break
			}
			panic(err)
		}
		transfers = append(transfers, models.NewTransfer(
			record.Key().ID.(int),
			record.Data().(*models.TransferData),
		))
	}

	return
}

//func renderContactUsers(w http.ResponseWriter, user, counterpartyUser models.AppUser) {
//
//}
//
//func renderCounterparty(w http.ResponseWriter, counterpartyUser models.AppUser, counterpartyContact models.Contact) {
//
//}
