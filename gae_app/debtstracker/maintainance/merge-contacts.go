package maintainance

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func mergeContactsHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	targetContactID, err := strconv.ParseInt(q.Get("target"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	sourceContacts := strings.Split(q.Get("source"), ",")
	sourceContactIDs := make([]int64, len(sourceContacts))
	for i, scID := range sourceContacts {
		if sourceContactIDs[i], err = strconv.ParseInt(scID, 10, 64); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}
	if err = mergeContacts(appengine.NewContext(r), targetContactID, sourceContactIDs...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Write([]byte("done"))
}

func mergeContacts(c context.Context, targetContactID int64, sourceContactIDs ...int64) (err error) {
	if len(sourceContactIDs) == 0 {
		panic("len(sourceContactIDs) == 0")
	}

	var (
		targetContact models.Contact
		user          models.AppUser
	)

	if targetContact, err = dal.Contact.GetContactByID(c, targetContactID); err != nil {
		if db.IsNotFound(err) && len(sourceContactIDs) == 1 {
			if targetContact, err = dal.Contact.GetContactByID(c, sourceContactIDs[0]); err != nil {
				return
			}
			targetContact.ID = targetContactID
			if err = dal.Contact.SaveContact(c, targetContact); err != nil {
				return
			}
		} else {
			return
		}
	}

	if user, err = dal.User.GetUserByID(c, targetContact.UserID); err != nil {
		return
	}

	for _, sourceContactID := range sourceContactIDs {
		if sourceContactID == targetContactID {
			err = fmt.Errorf("sourceContactID == targetContactID: %v", sourceContactID)
			return
		}
		var sourceContact models.Contact
		if sourceContact, err = dal.Contact.GetContactByID(c, sourceContactID); err != nil {
			if db.IsNotFound(err) {
				continue
			}
			return
		}
		if sourceContact.UserID != targetContact.UserID {
			err = fmt.Errorf("sourceContact.UserID != targetContact.UserID: %v != %v", sourceContact.UserID, targetContact.UserID)
			return
		}
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(sourceContactIDs))

	for _, sourceContactID := range sourceContactIDs {
		go func(sourceContactID int64) {
			if err2 := mergeContactTransfers(c, wg, targetContactID, sourceContactID); err2 != nil {
				log.Errorf(c, "failed to merge transfers for contact %v: %v", sourceContactID, err2)
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

	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if user, err = dal.User.GetUserByID(c, user.ID); err != nil {
			return
		}
		var contacts []models.UserContactJson
		userContacts := user.Contacts()
		targetContactBalance := targetContact.Balance()
		for _, contact := range userContacts {
			for _, sourceContactID := range sourceContactIDs {
				if contact.ID == sourceContactID {
					for currency, value := range contact.Balance() {
						targetContactBalance.Add(models.NewAmount(currency, value))
					}
					var sourceContact models.Contact
					if sourceContact, err = dal.Contact.GetContactByID(c, sourceContactID); err != nil {
						if db.IsNotFound(err) {
							err = nil
						} else {
							return
						}
					} else {
						targetContact.CountOfTransfers += sourceContact.CountOfTransfers
						if targetContact.LastTransferAt.Before(sourceContact.LastTransferAt) {
							targetContact.LastTransferAt = sourceContact.LastTransferAt
							targetContact.LastTransferID = sourceContact.LastTransferID
						}
						if sourceContact.CounterpartyCounterpartyID != 0 {
							var counterpartyContact models.Contact
							if counterpartyContact, err = dal.Contact.GetContactByID(c, sourceContact.CounterpartyCounterpartyID); err != nil {
								if db.IsNotFound(err) {
									err = nil
								} else {
									return
								}
							} else if counterpartyContact.CounterpartyCounterpartyID == sourceContactID {
								counterpartyContact.CounterpartyCounterpartyID = targetContactID
								if err = dal.Contact.SaveContact(c, counterpartyContact); err != nil {
									return
								}
							} else if counterpartyContact.CounterpartyCounterpartyID != 0 && counterpartyContact.CounterpartyCounterpartyID != targetContactID {
								err = fmt.Errorf(
									"data integrity issue : counterpartyContact(id=%v).CounterpartyCounterpartyID != sourceContactID: %v != %v",
									counterpartyContact.ID, counterpartyContact.CounterpartyCounterpartyID, sourceContactID)
								return
							}
						}
					}
					if err = dal.Contact.DeleteContact(c, sourceContactID); err != nil {
						return
					}
				} else {
					contacts = append(contacts, contact)
				}
			}
		}
		for i := range contacts {
			if contacts[i].ID == targetContactID {
				if err = contacts[i].SetBalance(targetContactBalance); err != nil {
					return
				}
				user.SetContacts(contacts)
				break
			}
		}

		if err = dal.User.SaveUser(c, user); err != nil {
			return
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		return errors.WithMessage(err, "failed to update user entity")
	}

	return
}

func mergeContactTransfers(c context.Context, wg *sync.WaitGroup, targetContactID int64, sourceContactID int64) (err error) {
	defer func() {
		wg.Done()
	}()
	transfersQ := datastore.NewQuery(models.TransferKind)
	transfersQ = transfersQ.Filter("BothCounterpartyIDs =", sourceContactID)
	transfers := transfersQ.Run(c)
	var (
		key      *datastore.Key
		transfer models.Transfer
	)
	for {
		transfer.TransferEntity = new(models.TransferEntity)
		if key, err = transfers.Next(transfer.TransferEntity); err != nil {
			if err == datastore.Done {
				err = nil
				break
			}
			log.Errorf(c, "Failed to get next transfer: %v", err)
		}
		transfer.ID = key.IntID()
		switch sourceContactID {
		case transfer.From().ContactID:
			transfer.From().ContactID = targetContactID
		case transfer.To().ContactID:
			transfer.To().ContactID = targetContactID
		}
		switch sourceContactID {
		case transfer.BothCounterpartyIDs[0]:
			transfer.BothCounterpartyIDs[0] = targetContactID
		case transfer.BothCounterpartyIDs[1]:
			transfer.BothCounterpartyIDs[1] = targetContactID
		}
		if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
			log.Errorf(c, "Failed to save transfer #%v: %v", transfer.ID, err)
		}
	}
	return
}
