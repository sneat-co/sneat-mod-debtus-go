package maintainance

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"github.com/strongo/nds"
	"google.golang.org/appengine/datastore"
)

type verifyTransfers struct {
	transfersAsyncJob
}

func (m *verifyTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	return m.startTransferWorker(c, counters, key, m.verifyTransfer)
}

func (m *verifyTransfers) verifyTransfer(c context.Context, counters *asyncCounters, transfer models.Transfer) (err error) {
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
	if err = m.verifyReturnsToTransferIDs(c, transfer, buf, counters); err != nil {
		log.Errorf(c, errors.WithMessage(err, "verifyReturnsToTransferIDs:transfer=%v").Error(), transfer.ID)
		return
	}
	if buf.Len() > 0 {
		log.Warningf(c, fmt.Sprintf("Transfer: %v, Created: %v\n", transfer.ID, transfer.DtCreated)+buf.String())
	}
	return
}

func (m *verifyTransfers) verifyTransferUsers(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters *asyncCounters) (err error) {
	for _, userID := range transfer.BothUserIDs {
		if userID != 0 {
			if _, err2 := facade.User.GetUserByID(c, userID); db.IsNotFound(err2) {
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

func (m *verifyTransfers) verifyTransferContacts(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters *asyncCounters) (err error) {
	for _, contactID := range transfer.BothCounterpartyIDs {
		if contactID != 0 {
			if _, err2 := facade.GetContactByID(c, contactID); db.IsNotFound(err2) {
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
			if user, err = facade.User.GetUserByID(c, toUse.UserID); err != nil {
				return changed, errors.WithMessage(err, "failed to get user by ID")
			}
			contactIDs := make([]int64, 0, user.ContactsCount)
			for _, c := range user.Contacts() {
				contactIDs = append(contactIDs, c.ID)
			}
			contacts, err := facade.GetContactsByIDs(c, contactIDs)
			if err != nil {
				return false, errors.WithMessage(err, fmt.Sprintf("failed to get contacts by IDs: %v", contactIDs))
			}
			for _, contact := range contacts {
				if contact.CounterpartyUserID == toFix.UserID {
					toFix.ContactID = contact.ID
					changed = true
					fmt.Fprintf(buf, "will assign ContactID=%v, ContactName=%v for UserID=%v, UserName=%v", contact.ID, contact.FullName(), from.UserID, from.UserName)
					break
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

func (*verifyTransfers) verifyTransferCurrency(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters *asyncCounters) (err error) {
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
			if transfer, err = facade.Transfers.GetTransferByID(c, transfer.ID); err != nil {
				return errors.WithMessage(err, "failed to get transfer by ID"+strconv.FormatInt(transfer.ID, 10))
			}
			if transfer.Currency != currency {
				transfer.Currency = currency
				if err = facade.Transfers.SaveTransfer(c, transfer); err != nil {
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

func (*verifyTransfers) verifyReturnsToTransferIDs(c context.Context, transfer models.Transfer, buf *bytes.Buffer, counters *asyncCounters) (err error) {
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
