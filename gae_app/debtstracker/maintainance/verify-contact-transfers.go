package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/captaincodeman/datastore-mapper"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
	"bytes"
	"github.com/sanity-io/litter"
	"strings"
	"github.com/strongo/app/db"
	"time"
	"github.com/strongo/app/gaedb"
	"encoding/json"
)

type verifyContactTransfers struct {
	contactsAsyncJob
}

func (m *verifyContactTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) error {
	return m.startContactWorker(c, counters, key, m.processContact)
}

func (m *verifyContactTransfers) processContact(c context.Context, counters *asyncCounters, contact models.Contact) (err error) {
	buf := new(bytes.Buffer)
	now := time.Now()
	hasError := false
	var (
		user           models.AppUser
		warningsCount  int
		transfers      []models.Transfer
		contactBalance models.Balance
	)

	defer func() {
		if hasError || warningsCount > 0 {
			var logFunc log.LogFunc
			if hasError {
				logFunc = log.Errorf
			} else {
				logFunc = log.Warningf
			}
			var userName, contactName string
			if user.AppUserEntity != nil {
				userName = user.FullName()
			}
			if contact.ContactEntity == nil {
				contactName = contact.FullName()
			}
			logFunc(c,
				fmt.Sprintf(
					`Contact(id=%v, name=%v): has %v warning, %v transfers
	User(%v): %v
	balance: %v
	`,
					contact.ID,
					contactName,
					warningsCount,
					len(transfers),
					user.ID,
					userName,
					litter.Sdump(contactBalance),
				)+ buf.String(),
			)
		}
	}()

	q := datastore.NewQuery(models.TransferKind) // TODO: Load outstanding transfer just for the specific contact & specific direction
	q = q.Filter("BothCounterpartyIDs =", contact.ID)
	q = q.Order("-DtCreated")
	transferEntities := make([]*models.TransferEntity, 0, contact.CountOfTransfers)
	var transferKeys []*datastore.Key
	if transferKeys, err = q.GetAll(c, &transferEntities); err != nil {
		log.Errorf(c, errors.WithMessage(err, "failed to load transfer").Error())
		return
	}
	transfers = make([]models.Transfer, len(transferEntities))
	for i, transfer := range transferEntities {
		transfers[i] = models.Transfer{ID: transferKeys[i].IntID(), TransferEntity: transfer}
	}
	models.ReverseTransfers(transfers)

	transfersByID := make(map[int64]models.Transfer, len(transfers))

	if len(transfers) != contact.CountOfTransfers {
		fmt.Fprintf(buf, "\tlen(transfers) != contact.CountOfTransfers: %v != %v\n", len(transfers), contact.CountOfTransfers)
		warningsCount += 1
	}

	if contact.CounterpartyCounterpartyID != 0 || contact.CounterpartyUserID != 0 { // Fixing names
		for _, transfer := range transfers {
			originalTransfer := transfer
			*originalTransfer.TransferEntity = *transfer.TransferEntity
			changed := false
			self := transfer.UserInfoByUserID(contact.UserID)
			counterparty := transfer.CounterpartyInfoByUserID(contact.UserID)

			if contact.CounterpartyUserID != 0 && counterparty.UserID == 0 {
				counterparty.UserID = contact.CounterpartyUserID
				changed = true
			}
			if counterparty.UserName == "" && counterparty.UserID != 0 {
				if user, err := dal.User.GetUserByID(c, counterparty.UserID); err != nil {
					log.Errorf(c, err.Error())
					return err
				} else {
					counterparty.UserName = user.FullName()
				}
			}

			if contact.CounterpartyCounterpartyID != 0 && self.ContactID == 0 {
				self.ContactID = contact.CounterpartyCounterpartyID
				changed = true
			}

			if self.ContactID != 0 && self.ContactName == "" {
				if counterpartyContact, err := dal.Contact.GetContactByID(c, self.ContactID); err != nil {
					log.Errorf(c, err.Error())
					return err
				} else {
					self.ContactName = counterpartyContact.FullName()
				}
			}

			if self.UserID != 0 && self.UserName == "" {
				if user, err := dal.User.GetUserByID(c, self.UserID); err != nil {
					log.Errorf(c, err.Error())
					return err
				} else {
					self.UserName = user.FullName()
				}
				changed = true
			}

			if changed {
				log.Warningf(c, "Fixing contact details for transfer %v: From:%v, To: %v\n\noriginal: %v\n\n new: %v", transfer.ID, litter.Sdump(transfer.From()), litter.Sdump(transfer.To()), litter.Sdump(originalTransfer), litter.Sdump(transfer))
				if err = dal.Transfer.SaveTransfer(c, transfer); err != nil {
					log.Errorf(c, errors.WithMessage(err, "failed to save transfer").Error())
					return
				}
			}
		}
	}

	loggedTransfers := make(map[int64]bool, len(transfers))

	logTransfer := func(transfer models.Transfer, padding int) {
		if _, ok := loggedTransfers[transfer.ID]; !ok {
			p := strings.Repeat("\t", padding)
			fmt.Fprintf(buf, p+"Transfer: %v\n", transfer.ID)
			fmt.Fprintf(buf, p+"\tCreated: %v\n", transfer.DtCreated)
			fmt.Fprintf(buf, p+"\tFrom(): userID=%v, contactID=%v\n", transfer.From().UserID, transfer.From().ContactID)
			fmt.Fprintf(buf, p+"\t  To(): userID=%v, contactID=%v\n", transfer.To().UserID, transfer.To().ContactID)
			fmt.Fprintf(buf, p+"\tAmount: %v\n", transfer.GetAmount())
			fmt.Fprintf(buf, p+"\tReturned: %v\n", transfer.AmountInCentsReturned)
			fmt.Fprintf(buf, p+"\tOutstanding: %v\n", transfer.GetOutstandingValue(time.Now()))
			fmt.Fprintf(buf, p+"\tIsReturn: %v\n", transfer.IsReturn)
			fmt.Fprintf(buf, p+"\tReturnTransferIDs: %v\n", transfer.ReturnTransferIDs)
			fmt.Fprintf(buf, p+"\tReturnToTransferIDs: %v\n", transfer.ReturnToTransferIDs)
			loggedTransfers[transfer.ID] = true
		}
	}

	transfersBalance := m.getTransfersBalance(transfers, contact.ID)

	verifyReturnIDs := func() (valid bool) {
		valid = true
		counters.Lock()
		for _, transfer := range transfersByID {
			for i, returnTransferID := range transfer.ReturnTransferIDs {
				if _, ok := transfersByID[returnTransferID]; ok {
					counters.Increment( "good_ReturnTransferID", 1)
				} else {
					valid = false
					logTransfer(transfer, 2)
					fmt.Fprintf(buf, "\t\tReturnTransferIDs[%d]: %v\n", i, returnTransferID)
					counters.Increment( "wrong_ReturnTransferID", 1)
					warningsCount += 1
				}
			}
			for i, returnToTransferID := range transfer.ReturnToTransferIDs {
				if _, ok := transfersByID[returnToTransferID]; ok {
					counters.Increment( "good_ReturnToTransferID", 1)
				} else {
					valid = false
					logTransfer(transfer, 2)
					fmt.Fprintf(buf, "\t\tReturnToTransferIDs[%d]: %v\n", i, returnToTransferID)
					counters.Increment( "wrong_ReturnToTransferID", 1)
					warningsCount += 1
				}
			}
		}
		counters.Unlock()
		return
	}

	m.verifyTotals(buf, counters, contact, transfersBalance)

	outstandingIsValid, outstandingWarningsCount := m.verifyOutstanding(c, 1, buf, contact, transfers)
	warningsCount += outstandingWarningsCount
	if !outstandingIsValid {
		//rollingBalance := make(models.Balance, len(transfersBalance)+1)
		transfersByCurrency, transfersToSave := m.fixTransfers(c, now, buf, contact, transfers)

		for currency, currencyTransfers := range transfersByCurrency {
			fmt.Fprintf(buf, "\tcurrency: %v - %d transfers\n", currency, len(currencyTransfers))
		}

		if valid, _ := m.verifyOutstanding(c, 2, buf, contact, transfers); !valid {
			fmt.Fprint(buf, "Outstandings are invalid after fix")
		} else if valid, _ = m.verifyTotals(buf, counters, contact, transfersBalance); !valid {
			fmt.Fprint(buf, "Totals are invalid after fix")
		} else if valid = verifyReturnIDs(); !valid {
			fmt.Fprint(buf, "Return IDs are invalid after fix")
		} else {
			fmt.Fprintf(buf, "%v transfers to save!\n", len(transfersToSave))
			m.logTransfers(transfers, 1, true)
			entitiesToSave := make([]db.EntityHolder, 0, len(transfersToSave))
			for id, transfer := range transfersToSave {
				entitiesToSave = append(entitiesToSave, &models.Transfer{ID: id, TransferEntity: transfer})
			}
			gaedb.LoggingEnabled = true
			if err = dal.DB.UpdateMulti(c, entitiesToSave); err != nil {
				gaedb.LoggingEnabled = false
				fmt.Fprintf(buf, "ERROR: failed to save transfers: "+err.Error())
				hasError = true
				return
			}
			gaedb.LoggingEnabled = false
			fmt.Fprintf(buf, "SAVED %v transfers!\n", len(entitiesToSave))
		}
	}

	if outstandingIsValid {
		if user, err = dal.User.GetUserByID(c, contact.UserID); err != nil {
			log.Errorf(c, errors.WithMessage(err, fmt.Sprintf("Contact(%v): ", contact.ID)+"user not found by ID").Error())
			return
		}
	}

	if !outstandingIsValid || !contact.Balance().Equal(user.ContactByID(contact.ID).Balance()) {
		if contact, user, err = m.updateContactAndUser(c, buf, contact.ID, transfers); err != nil {
			return
		}
	}

	if warningsCount == 0 {
		counters.Increment( "good_contacts", 1)
		//log.Infof(c, contactPrefix + "is OK, %v transfers", len(transfers))
	} else {
		counters.Lock()
		counters.Increment("bad_contacts", 1)
		counters.Increment("warnings", int64(warningsCount))
		counters.Unlock()

		contactBalance := contact.Balance()

		if len(contactBalance) == 0 {
			contactBalance = nil
		}
	}
	return nil
}

func (m *verifyContactTransfers) updateContactAndUser(c context.Context, buf *bytes.Buffer, contactID int64, transfers []models.Transfer) (contact models.Contact, user models.AppUser, err error) {
	if contactID == 0 {
		err = errors.New("*verifyContactTransfers.updateContactAndUser(): contactID == 0")
		return
	}
	transfersBalance := m.getTransfersBalance(transfers, contactID)
	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if contact, err = dal.Contact.GetContactByID(c, contactID); err != nil {
			return
		}
		if contactBalance := contact.Balance(); !contactBalance.Equal(transfersBalance) {
			if err = contact.SetBalance(transfersBalance); err != nil {
				return
			}
			fmt.Fprintf(buf, "contact balance update from transfers\nwas: %v\nnew: %v", contactBalance, transfersBalance)
			if err = dal.Contact.SaveContact(c, contact); err != nil {
				return
			}
		}
		if user, err = dal.User.GetUserByID(c, contact.UserID); err != nil {
			return
		}
		userChanged := user.AddOrUpdateContact(contact)
		userContacts := user.Contacts()
		for i, uc := range userContacts {
			if uc.ID == contact.ID {
				ucChanged := false
				if (uc.BalanceJson == nil && contact.BalanceJson != "") || (uc.BalanceJson != nil && string(*uc.BalanceJson) != contact.BalanceJson) {
					balanceJson := json.RawMessage(contact.BalanceJson)
					uc.BalanceJson = &balanceJson
					ucChanged = true
				}
				if len(transfers) > 0 {
					lastTransfer := transfers[len(transfers)-1]
					if uc.Transfers == nil {
						uc.Transfers = &models.UserContactTransfersInfo{}
						ucChanged = true
					}
					if uc.Transfers.Last.ID != lastTransfer.ID {
						uc.Transfers.Last.ID = lastTransfer.ID
						ucChanged = true
					}
					if !uc.Transfers.Last.At.Equal(lastTransfer.DtCreated) {
						uc.Transfers.Last.At = lastTransfer.DtCreated
						ucChanged = true
					}
					if uc.Transfers.Count != len(transfers) {
						uc.Transfers.Count = len(transfers)
						ucChanged = true
					}
					// TODO: check outstanding without interest
				}
				if ucChanged {
					userContacts[i] = uc
					user.SetContacts(userContacts)
					userChanged = true
				}
				break
			}
		}
		userTotalBalance := user.Balance()
		if userContactsBalance := user.TotalBalanceFromContacts(); !userContactsBalance.Equal(userTotalBalance) {
			if err = user.SetBalance(userContactsBalance); err != nil {
				return
			}
			fmt.Fprintln(buf, "user total balance update from contacts\nwas: %v\nnew: %v", userTotalBalance, userContactsBalance)
		}
		if userChanged {
			if err = dal.User.SaveUser(c, user); err != nil {
				return
			}
		}
		return
	}, db.CrossGroupTransaction)
	if err != nil {
		log.Errorf(c, "failed to updated contact & user: %v", err)
		return
	}
	return
}

func (m *verifyContactTransfers) verifyTotals(buf *bytes.Buffer, counters *asyncCounters, contact models.Contact, transfersBalance models.Balance) (valid bool, warningsCount int) {
	valid = true
	contactBalance := contact.Balance()
	for currency, transfersTotal := range transfersBalance {
		if contactTotal := contactBalance[currency]; contactTotal != transfersTotal {
			valid = false
			fmt.Fprintf(buf, "currency %v: transfersTotal != contactTotal: %v != %v\n", currency, transfersTotal, contactTotal)
			warningsCount += 1
		}
		delete(contactBalance, currency)
	}
	for currency, contactTotal := range contactBalance {
		if contactTotal == 0 {
			counters.Increment( "zero_balance", 1)
			fmt.Fprintf(buf, "\t0 value for currency %v\n", currency)
			warningsCount += 1
		} else {
			counters.Increment( "no_transfers_for_non_zero_balance", 1)
			fmt.Fprintf(buf, "\tno transfers found for %v=%v\n", currency, contactTotal)
			warningsCount += 1
		}
	}
	return
}

func (verifyContactTransfers) getTransfersBalance(transfers []models.Transfer, contactID int64) (totalBalance models.Balance) {
	totalBalance = make(models.Balance)
	for _, transfer := range transfers {
		direction := transfer.DirectionForContact(contactID)
		switch direction {
		case models.TransferDirectionUser2Counterparty:
			totalBalance[transfer.Currency] += transfer.AmountInCents
		case models.TransferDirectionCounterparty2User:
			totalBalance[transfer.Currency] -= transfer.AmountInCents
		default:
			panic(fmt.Sprintf("transfer.DirectionForContact(%v): %v", contactID, direction))
		}
	}
	for c, v := range totalBalance {
		if v == 0 {
			delete(totalBalance, c)
		}
	}
	return
}

func (verifyContactTransfers) getTransfersOutstanding(transfers []models.Transfer, contactID int64, retortTime time.Time) (outstandingBalance models.Balance) {
	outstandingBalance = make(models.Balance)

	for _, transfer := range transfers {
		//logTransfer(transfer, 1)
		direction := transfer.DirectionForContact(contactID)
		switch direction {
		case models.TransferDirectionUser2Counterparty:
			outstandingBalance[transfer.Currency] += transfer.GetOutstandingValue(retortTime)
		case models.TransferDirectionCounterparty2User:
			outstandingBalance[transfer.Currency] -= transfer.GetOutstandingValue(retortTime)
		default:
			panic(fmt.Sprintf("transfer.DirectionForContact(%v): %v", contactID, direction))
		}
	}
	for c, v := range outstandingBalance {
		if v == 0 {
			delete(outstandingBalance, c)
		}
	}
	return
}

func (m *verifyContactTransfers) verifyOutstanding(c context.Context, iteration int, buf *bytes.Buffer, contact models.Contact, transfers []models.Transfer) (valid bool, warningsCount int) {
	valid = true
	transfersOutstanding := m.getTransfersOutstanding(transfers, contact.ID, time.Now())
	transfersBalance := m.getTransfersBalance(transfers, contact.ID)
	for currency, balanceTotal := range transfersBalance {
		if outstandingTotal := transfersOutstanding[currency]; outstandingTotal == balanceTotal {
			fmt.Fprintf(buf, "\t%v: balanceTotal == outstandingTotal: %v\n", currency, balanceTotal)
		} else {
			valid = false
			fmt.Fprintf(buf, "\tcurrency %v: balanceTotal != outstandingTotal: %v != %v\n", currency, balanceTotal, outstandingTotal)
			warningsCount += 1
		}
		delete(transfersOutstanding, currency)
	}
	fmt.Fprintf(buf, "\tverifyOutstanding(iteration=%v): valid=%v\n", iteration, valid)
	return
}

func (m *verifyContactTransfers) logTransfers(transfers []models.Transfer, padding int, reset bool) {
	//if reset {
	//	loggedTransfers = make(map[int64]bool, len(transfers))
	//}
	//for _, transfer := range transfers {
	//	logTransfer(transfer, 1)
	//}
}

func (m *verifyContactTransfers) fixTransfers(c context.Context, now time.Time, buf *bytes.Buffer, contact models.Contact, transfers []models.Transfer) (
	transfersByCurrency map[models.Currency][]models.Transfer,
	transfersToSave map[int64]*models.TransferEntity,
) {
	fmt.Fprintf(buf, "Will try to fix %d transfers:\n", len(transfers))
	m.logTransfers(transfers, 1, true)

	transfersByCurrency = make(map[models.Currency][]models.Transfer)

	transfersToSave = make(map[int64]*models.TransferEntity)

	for _, transfer := range transfers {
		if transfer.AmountInCentsReturned != 0 {
			transfer.AmountInCentsReturned = 0
			transfersToSave[transfer.ID] = transfer.TransferEntity
		}
		if len(transfer.ReturnTransferIDs) != 0 {
			transfer.ReturnTransferIDs = []int64{}
			transfersToSave[transfer.ID] = transfer.TransferEntity
		}
		if len(transfer.ReturnToTransferIDs) != 0 {
			transfer.ReturnToTransferIDs = []int64{}
			transfersToSave[transfer.ID] = transfer.TransferEntity
		}
		amountToAssign := transfer.GetAmount().Value
		for _, previousTransfer := range transfersByCurrency[transfer.Currency] {
			if previousTransfer.IsOutstanding && previousTransfer.IsReverseDirection(transfer.TransferEntity) {
				previousTransfer.ReturnTransferIDs = append(previousTransfer.ReturnTransferIDs, transfer.ID)
				transfer.ReturnToTransferIDs = append(transfer.ReturnToTransferIDs, previousTransfer.ID)
				transfersToSave[previousTransfer.ID] = previousTransfer.TransferEntity
				if previousTransferOutstandingValue := previousTransfer.GetOutstandingValue(now); amountToAssign <= previousTransferOutstandingValue {
					previousTransfer.AmountInCentsReturned += amountToAssign
					amountToAssign = 0
					break
				} else /* previousTransfer.AmountInCentsOutstanding < amountToAssign */ {
					amountToAssign -= previousTransferOutstandingValue
					previousTransfer.AmountInCentsReturned += previousTransferOutstandingValue
					previousTransfer.IsOutstanding = false
				}
			}
		}
		transfer.IsReturn = len(transfer.ReturnToTransferIDs) > 0
		if transfer.IsOutstanding = amountToAssign != 0; transfer.IsOutstanding {
			transfer.AmountInCentsReturned = transfer.AmountInCents - amountToAssign
			transfersToSave[transfer.ID] = transfer.TransferEntity
		}
		transfersByCurrency[transfer.Currency] = append(transfersByCurrency[transfer.Currency], transfer)
	}
	return
}
