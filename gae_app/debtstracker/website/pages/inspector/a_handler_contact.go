package inspector

import (
	"fmt"
	"net/http"
	"strconv"
	//"sync"

	"sync"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type contactPage struct {
}

func (h contactPage) contactPageHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	contactID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)

	var contact models.Contact

	if contact, err = facade.GetContactByID(c, contactID); err != nil {
		fmt.Fprint(w, err)
		return
	}

	//var user, counterpartyUser models.AppUser
	//var counterpartyContact models.Contact
	//
	var transfers []models.Transfer
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		if transfers, err = h.verifyTransfers(c, contactID); err != nil {
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
	//	if counterpartyContact, err = facade.GetContactByID(c, contact.CounterpartyCounterpartyID); err != nil {
	//		return
	//	}
	//}

	RenderContactPage(contact, w)

	//renderContactUsers(w, user, counterpartyUser)

}

func (contactPage) verifyTransfers(c context.Context, contactID int64) (
	transfers []models.Transfer, err error,
) {
	query := datastore.NewQuery(models.TransferKind).Filter("BothCounterpartyIDs=", contactID)

	iterator := query.Run(c)

	for {
		transferEntity := new(models.TransferEntity)
		var key *datastore.Key
		if key, err = iterator.Next(transferEntity); err != nil {
			if err == datastore.Done {
				break
			}
			panic(err)
		}
		transfers = append(transfers, models.NewTransfer(key.IntID(), transferEntity))
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
