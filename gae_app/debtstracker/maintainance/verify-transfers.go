package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/qedus/nds"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"net/http"
	"strings"
	"strconv"
	"bytes"
	"sync"
	"github.com/strongo/app/gaedb"
	"github.com/pkg/errors"
)

type verifyTransfers struct {
	sync.Mutex
	wg     sync.WaitGroup
	entity *models.TransferEntity
}

var _ mapper.SliceLifecycle = (*verifyTransfers)(nil)

func (m *verifyTransfers) SliceStarted(c context.Context, id string, namespace string, shard, slice int) {
	gaedb.LoggingEnabled = false
}

// SliceStarted is called when a mapper job for an individual slice of a
// shard within a namespace is completed
func (m *verifyTransfers) SliceCompleted(c context.Context, id string, namespace string, shard, slice int) {
	log.Debugf(c, "Awaiting completion...")
	m.wg.Wait()
	log.Debugf(c, "Processing completed.")
	gaedb.LoggingEnabled = true
}

func (m *verifyTransfers) Query(r *http.Request) (*mapper.Query, error) {
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user"), 10, 64)
	query := mapper.NewQuery(models.TransferKind)
	if userID != 0 {
		query = query.Filter("BothUserIDs =", userID)
	}
	return query, nil
}

func (m *verifyTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	transferEntity := *m.entity
	m.wg.Add(1)
	go m.verifyTransfer(c, counters, models.Transfer{ID: key.IntID(), TransferEntity: &transferEntity})
	return
}

func (m *verifyTransfers) verifyTransfer(c context.Context, counters mapper.Counters, transfer models.Transfer) {
	defer m.wg.Done()
	var err error
	buf := new(bytes.Buffer)
	if err = m.verifyTransferUsers(c, transfer, buf, counters); err != nil {
		log.Errorf(c, errors.WithMessage(err, "verifyTransferUsers:transfer=%v").Error(), transfer.ID)
		return
	}
	if err = m.verifyTransferContacts(c, transfer, buf, counters); err != nil {
		log.Errorf(c, errors.WithMessage(err, "verifyTransferContacts:transfer=%v").Error(), transfer.ID)
		return
	}
	if err = m.verifyTransferCurrency(c, transfer, buf, counters); err != nil {
		log.Errorf(c, errors.WithMessage(err, "verifyTransferCurrency:transfer=%v").Error(), transfer.ID)
		return
	}
	if err = m.verifyReturnsTransferIDs(c, transfer, buf, counters); err != nil {
		log.Errorf(c, errors.WithMessage(err, "verifyReturnsTransferIDs:transfer=%v").Error(), transfer.ID)
		return
	}
	if err = m.verifyReturnsToTransferIDs(c, transfer, buf, counters); err != nil {
		log.Errorf(c, errors.WithMessage(err, "verifyReturnsToTransferIDs:transfer=%v").Error(), transfer.ID)
		return
	}
	if buf.Len() > 0 {
		log.Warningf(c, fmt.Sprintf("Transfer: %v, Created: %v\n", transfer.ID, transfer.DtCreated)+buf.String())
	}
}

// JobStarted is called when a mapper job is started
func (m *verifyTransfers) JobStarted(c context.Context, id string) {

}

// JobStarted is called when a mapper job is completed
func (m *verifyTransfers) JobCompleted(c context.Context, id string) {
	logJobCompletion(c, id)
}

func (m *verifyTransfers) Make() interface{} {
	m.entity = new(models.TransferEntity)
	return m.entity
}

func (m *verifyTransfers) verifyTransferUsers(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters mapper.Counters) (err error) {
	for _, userID := range transfer.BothUserIDs {
		if userID != 0 {
			if _, err2 := dal.User.GetUserByID(c, userID); db.IsNotFound(err2) {
				counters.Increment(fmt.Sprintf("User:%d", userID), 1)
				fmt.Fprintf(buf, "Unknown user %d\n", userID)
			} else if err2 != nil {
				err = errors.WithMessage(err2, fmt.Sprintf("failed to get user by ID=%v", userID))
				return
			}
		}
	}
	return
}

func (m *verifyTransfers) verifyTransferContacts(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters mapper.Counters) (err error) {
	for _, contactID := range transfer.BothCounterpartyIDs {
		if contactID != 0 {
			if _, err2 := dal.Contact.GetContactByID(c, contactID); db.IsNotFound(err2) {
				counters.Increment(fmt.Sprintf("Contact:%d", contactID), 1)
				fmt.Fprintf(buf, "Unknown contact %d\n", contactID)
			} else if err2 != nil {
				err = errors.WithMessage(err2, fmt.Sprintf("failed to get contact by ID=%v", contactID))
				return
			}
		}
	}
	from := transfer.From()
	to := transfer.To()

	if from.UserID != 0 && to.UserID != 0 {
		fixContactID := func(toFix, toUse *models.TransferCounterpartyInfo) (changed bool, err error) {
			if toFix.ContactID != 0 {
				panic("toFix.ContactID != 0")
			}
			var user models.AppUser
			if user, err = dal.User.GetUserByID(c, toUse.UserID); err != nil {
				return changed, errors.WithMessage(err, "failed to get user by ID")
			}
			contactIDs := make([]int64, 0, user.ContactsCount)
			for _, c := range user.Contacts() {
				contactIDs = append(contactIDs, c.ID)
			}
			contacts, err := dal.Contact.GetContactsByIDs(c, contactIDs)
			if err != nil {
				return false, errors.WithMessage(err, fmt.Sprintf("failed to get contacts by IDs: %v", contactIDs))
			}
			for _, contact := range contacts {
				if contact.CounterpartyUserID == toFix.UserID {
					toFix.ContactID = contact.ID
					changed = true
					break
					fmt.Fprintf(buf, "will assign ContactID=%v, ContactName=%v for UserID=%v, UserName=%v", contact.ID, contact.FullName(), from.UserID, from.UserName)
				}
			}
			return changed, nil
		}
		var transferChanged, changed bool

		if from.ContactID == 0 {
			if changed, err = fixContactID(from, to); err != nil {
				return
			} else if changed {
				transferChanged = transferChanged || changed
			}
		}
		if to.ContactID == 0 {
			if changed, err = fixContactID(to, from); err != nil {
				return
			} else if changed {
				transferChanged = transferChanged || changed
			}
		}
	}
	return nil
}

func (_ *verifyTransfers) verifyTransferCurrency(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters mapper.Counters) (err error) {
	var currency models.Currency
	if transfer.Currency == models.Currency("euro") {
		currency = models.Currency("EUR")
	} else if len(transfer.Currency) == 3 {
		if v2 := models.Currency(strings.ToUpper(string(transfer.Currency))); v2 != transfer.Currency && v2.IsMoney() {
			currency = v2
		}
	}
	if currency != "" {
		if err = nds.RunInTransaction(c, func(c context.Context) error {
			if transfer, err = dal.Transfer.GetTransferByID(c, transfer.ID); err != nil {
				return errors.WithMessage(err, "failed to get transfer by ID" + strconv.FormatInt(transfer.ID, 10))
			}
			if transfer.Currency != currency {
				transfer.Currency = currency
				if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
					return errors.WithMessage(err, "failed to save transfer")
				}
				fmt.Fprintf(buf, "Currency fixed: %d\n", transfer.ID)
			}
			return nil
		}, nil); err != nil {
			return err
		}
	}
	return
}

func (_ *verifyTransfers) verifyReturnsTransferIDs(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters mapper.Counters) (err error) {
	if len(transfer.ReturnTransferIDs) == 0 {
		return
	}
	var returnTransfers []models.Transfer
	if returnTransfers, err = dal.Transfer.GetTransfersByID(c, transfer.ReturnTransferIDs); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("failed to get transfers by IDs: %v", transfer.ReturnTransferIDs))
	}
	for _, returnTransfer := range returnTransfers {
		if transfer.From().ContactID != returnTransfer.To().ContactID {
			fmt.Fprintf(buf, "returnTransfer(id=%v).To().ContactID != From().ContactID: %v != %v\n", returnTransfer.ID, returnTransfer.To().ContactID, transfer.From().ContactID)
		}
		if transfer.To().ContactID != returnTransfer.From().ContactID {
			fmt.Fprintf(buf, "returnTransfer(id=%v).From().ContactID != To().ContactID: %v != %v\n", returnTransfer.ID, returnTransfer.From().ContactID, transfer.To().ContactID)
		}
	}
	return
}

func (_ *verifyTransfers) verifyReturnsToTransferIDs(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters mapper.Counters) (err error) {
	if len(transfer.ReturnToTransferIDs) == 0 {
		return
	}
	var returnToTransfers []models.Transfer
	if returnToTransfers, err = dal.Transfer.GetTransfersByID(c, transfer.ReturnToTransferIDs); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("failed to get transfers by IDs: %v", transfer.ReturnToTransferIDs))
	}
	for _, returnToTransfer := range returnToTransfers {
		if transfer.From().ContactID != returnToTransfer.To().ContactID {
			fmt.Fprintf(buf, "returnToTransfer(id=%v).To().ContactID != From().ContactID: %v != %v\n", returnToTransfer.ID, returnToTransfer.To().ContactID, transfer.From().ContactID)
		}
		if transfer.To().ContactID != returnToTransfer.From().ContactID {
			fmt.Fprintf(buf, "returnToTransfer(id=%v).From().ContactID != To().ContactID: %v != %v\n", returnToTransfer.ID, returnToTransfer.From().ContactID, transfer.To().ContactID)
		}
	}
	return
}
