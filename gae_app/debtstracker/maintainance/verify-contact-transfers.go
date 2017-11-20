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
)

type verifyContactTransfers struct {
	contactsAsyncJob
}

func (m *verifyContactTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) error {
	return m.startContactWorker(c, counters, key, m.processContact)
}

func (m *verifyContactTransfers) processContact(c context.Context, counters *asyncCounters, contact models.Contact) (err error) {
	log.Debugf(c, "processContact(contact.ID=%v)", contact.ID)
	buf := new(bytes.Buffer)
	now := time.Now()
	hasError := false
	var (
		user           models.AppUser
		warningsCount  int
		transfers      []models.Transfer
	)
	contactBalance := contact.Balance()

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
					"Contact(id=%v, name=%v): has %v warning, %v transfers\n"+
						"\tcontact.Balance: %v\n"+
						"\tUser(id=%v, name=%v)",
					contact.ID,
					contactName,
					warningsCount,
					len(transfers),
					litter.Sdump(contactBalance),
					contact.UserID,
					userName,
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
		transfers[i] = models.Transfer{IntegerID: db.NewIntID(transferKeys[i].IntID()), TransferEntity: transfer}
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
			if transfer.HasInterest() {
				fmt.Fprintf(buf, p+"\tInterest: %v @ %v%%/%v_days, min=%v, grace=%v",
					transfer.InterestType, transfer.InterestPercent, transfer.InterestPeriod,
					transfer.InterestMinimumPeriod, transfer.InterestGracePeriod)
			}
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
					counters.Increment("good_ReturnTransferID", 1)
				} else {
					valid = false
					fmt.Fprintf(buf, "\t\tReturnTransferIDs[%d]: %v\n", i, returnTransferID)
					counters.Increment("wrong_ReturnTransferID", 1)
					warningsCount += 1
				}
			}
			for i, returnToTransferID := range transfer.ReturnToTransferIDs {
				if _, ok := transfersByID[returnToTransferID]; ok {
					counters.Increment("good_ReturnToTransferID", 1)
				} else {
					valid = false
					fmt.Fprintf(buf, "\t\tReturnToTransferIDs[%d]: %v\n", i, returnToTransferID)
					counters.Increment("wrong_ReturnToTransferID", 1)
					warningsCount += 1
				}
			}
		}
		counters.Unlock()
		return
	}

	var lastTransfer models.Transfer

	if len(transfers) > 0 {
		lastTransfer = transfers[len(transfers)-1]
	}

	var needsFixingContactOrUser bool

	if valid, warnsCount := m.assertTotals(buf, counters, contact, transfersBalance); !valid {
		needsFixingContactOrUser = true
		warningsCount += warnsCount
	} else {
		warningsCount += warnsCount
	}

	outstandingIsValid, outstandingWarningsCount := m.verifyOutstanding(c, 1, buf, contactBalance, transfersBalance)
	warningsCount += outstandingWarningsCount
	if !outstandingIsValid {
		//rollingBalance := make(models.Balance, len(transfersBalance)+1)
		transfersByCurrency, transfersToSave := m.fixTransfers(c, now, buf, contact, transfers)

		for currency, currencyTransfers := range transfersByCurrency {
			fmt.Fprintf(buf, "\tcurrency: %v - %d transfers\n", currency, len(currencyTransfers))
		}

		if valid, _ := m.verifyOutstanding(c, 2, buf, contactBalance, transfersBalance); !valid {
			fmt.Fprint(buf, "Outstandings are invalid after fix!\n")
			needsFixingContactOrUser = true
		} else if valid, _ = m.assertTotals(buf, counters, contact, transfersBalance); !valid {
			fmt.Fprint(buf, "Totals are invalid after fix!\n")
		} else if valid = verifyReturnIDs(); !valid {
			fmt.Fprint(buf, "Return IDs are invalid after fix!\n")
		} else {
			fmt.Fprintf(buf, "%v transfers to save!\n", len(transfersToSave))
			entitiesToSave := make([]db.EntityHolder, 0, len(transfersToSave))
			for id, transfer := range transfersToSave {
				entitiesToSave = append(entitiesToSave, &models.Transfer{IntegerID: db.NewIntID(id), TransferEntity: transfer})
			}
			gaedb.LoggingEnabled = true
			if err = dal.DB.UpdateMulti(c, entitiesToSave); err != nil {
				gaedb.LoggingEnabled = false
				fmt.Fprintf(buf, "ERROR: failed to save transfers: %v\n", err)
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


	if !outstandingIsValid || !contactBalance.Equal(user.ContactByID(contact.ID).Balance()) || !contactBalance.Equal(transfersBalance) {
		needsFixingContactOrUser = true
	}

	if !needsFixingContactOrUser && contact.CounterpartyCounterpartyID != 0 {
		var counterpartyContact models.Contact
		if counterpartyContact, err = dal.Contact.GetContactByID(c, contact.CounterpartyCounterpartyID); err != nil {
			return
		}
		fmt.Fprintf(buf,"contact.Balance(): %v\n", contact.Balance())
		fmt.Fprintf(buf,"counterpartyContact.Balance(): %v\n", contact.Balance())
		if !counterpartyContact.GetTransfersInfo().Equal(contact.GetTransfersInfo()) || !counterpartyContact.Balance().Equal(transfersBalance.Reversed()) {
			needsFixingContactOrUser = true
		}
	} else {
		fmt.Fprintf(buf, "needsFixingContactOrUser: %v, contact.CounterpartyCounterpartyID: %v", needsFixingContactOrUser, contact.CounterpartyCounterpartyID)
	}

	if needsFixingContactOrUser {
		for _, transfer := range transfers {
			logTransfer(transfer, 1)
		}
		if contact, user, err = m.fixContactAndUser(c, buf, counters, contact.ID, transfersBalance, len(transfers), lastTransfer); err != nil {
			return
		}
	}

	if warningsCount == 0 {
		counters.Increment("good_contacts", 1)
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

func (m *verifyContactTransfers) assertTotals(buf *bytes.Buffer, counters *asyncCounters, contact models.Contact, transfersBalance models.Balance) (valid bool, warningsCount int) {
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
			counters.Increment("zero_balance", 1)
			fmt.Fprintf(buf, "\t0 value for currency %v\n", currency)
			warningsCount += 1
		} else {
			counters.Increment("no_transfers_for_non_zero_balance", 1)
			fmt.Fprintf(buf, "\tno transfers found for %v=%v\n", currency, contactTotal)
			warningsCount += 1
		}
	}
	return
}

func (m *verifyContactTransfers) fixContactAndUser(c context.Context, buf *bytes.Buffer, counters *asyncCounters, contactID int64, transfersBalance models.Balance, transfersCount int, lastTransfer models.Transfer) (contact models.Contact, user models.AppUser, err error) {
	if err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if contact, user, err = m.fixContactAndUserWithinTransaction(c, buf, counters, contactID, transfersBalance, transfersCount, lastTransfer); err != nil {
			return
		}
		if contact.CounterpartyCounterpartyID != 0 {
			if _, _, err = m.fixContactAndUserWithinTransaction(c, buf, counters, contact.CounterpartyCounterpartyID, transfersBalance.Reversed(), transfersCount, lastTransfer); err != nil {
				return
			}
		}
		return
	}, db.CrossGroupTransaction); err != nil {
		return
	}
	return
}

func (m *verifyContactTransfers) fixContactAndUserWithinTransaction(c context.Context, buf *bytes.Buffer, counters *asyncCounters, contactID int64, transfersBalance models.Balance, transfersCount int, lastTransfer models.Transfer) (contact models.Contact, user models.AppUser, err error) {
	fmt.Fprintf(buf,"Fixing contact %v...\n", contactID)
	if contact, err = dal.Contact.GetContactByID(c, contactID); err != nil {
		return
	}
	changed := false
	if lastTransfer.TransferEntity != nil && lastTransfer.ID != 0 {
		if contact.LastTransferAt.Before(lastTransfer.DtCreated) {
			fmt.Fprintf(buf, "\tcontact.LastTransferAt changed from %v to %v\n", contact.LastTransferID, lastTransfer.DtCreated)
			contact.LastTransferAt = lastTransfer.DtCreated

			if contact.LastTransferID != lastTransfer.ID {
				fmt.Fprintf(buf, "\tcontact.LastTransferID changed from %v to %v\n", contact.LastTransferID, lastTransfer.ID)
				contact.LastTransferID = lastTransfer.ID
			}
			changed = true
		}
	}
	if contact.CountOfTransfers < transfersCount {
		fmt.Fprintf(buf, "\tcontact.CountOfTransfers changed from %v to %v\n", contact.CountOfTransfers, transfersCount)
		contact.CountOfTransfers = transfersCount
		changed = true
	}
	if !contact.Balance().Equal(transfersBalance) {
		if err = contact.SetBalance(transfersBalance); err != nil {
			return
		}
		changed = true
	}
	if changed {
		if err = dal.Contact.SaveContact(c, contact); err != nil {
			return
		}
		//var user models.AppUser
		if user, err = dal.User.GetUserByID(c, contact.UserID); err != nil {
			return
		}
		userContacts := user.Contacts()
		userChanged := false
		for i := range userContacts {
			if userContacts[i].ID == contact.ID {
				if !userContacts[i].Balance().Equal(transfersBalance) {
					userContacts[i].SetBalance(transfersBalance)
					user.SetContacts(userContacts)
					userChanged = true
				}
				userTransferInfo, contactTransferInfo := userContacts[i].Transfers, contact.GetTransfersInfo()
				if !userTransferInfo.Equal(contactTransferInfo) {
					userContacts[i].Transfers = contactTransferInfo
					userChanged = true
				}
				goto contactFound
			}
		}
		// Contact not found
		userChanged = user.AddOrUpdateContact(contact) || userChanged
	contactFound:
		userTotalBalance := user.Balance()
		if userContactsBalance := user.TotalBalanceFromContacts(); !userContactsBalance.Equal(userTotalBalance) {
			if err = user.SetBalance(userContactsBalance); err != nil {
				return
			}
			userChanged = true
			fmt.Fprintln(buf, "user total balance update from contacts\nwas: %v\nnew: %v", userTotalBalance, userContactsBalance)
		}
		if userChanged {
			if err = dal.User.SaveUser(c, user); err != nil {
				return
			}
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

func (m *verifyContactTransfers) verifyOutstanding(c context.Context, iteration int, buf *bytes.Buffer, contactBalance models.Balance, transfersBalance models.Balance) (valid bool, warningsCount int) {
	fmt.Fprintf(buf, "\tverifyOutstanding(iteration=%v):\n", iteration)
	valid = true

	for currency, contactTotal := range contactBalance {
		if transfersTotal := transfersBalance[currency]; transfersTotal == contactTotal {
			fmt.Fprintf(buf, "\t\tcurrency %v: contactBalance == transfersTotal: %v\n", currency, contactTotal)
		} else {
			valid = false
			fmt.Fprintf(buf, "\t\tcurrency %v: contactBalance != transfersTotal: %v != %v\n", currency, contactTotal, transfersTotal)
			warningsCount += 1
		}
		//delete(transfersOutstanding, currency)
	}
	fmt.Fprintf(buf, "\tverifyOutstanding(iteration=%v) => valid=%v\n", iteration, valid)

	return
}

func (m *verifyContactTransfers) fixTransfers(c context.Context, now time.Time, buf *bytes.Buffer, contact models.Contact, transfers []models.Transfer) (
	transfersByCurrency map[models.Currency][]models.Transfer,
	transfersToSave map[int64]*models.TransferEntity,
) {
	fmt.Fprintf(buf, "fixTransfers()\n")

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
