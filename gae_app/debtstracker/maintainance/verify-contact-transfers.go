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
)

type verifyContactTransfers struct {
	contactsAsyncJob
}

func (m *verifyContactTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) error {
	return m.startProcess(c, func() func(){
		contactEntity := *m.entity
		contact := models.Contact{ID: key.IntID(), ContactEntity: &contactEntity}
		return func() { m.processContact(c, contact, counters) }
	})
}

func (m *verifyContactTransfers) processContact(c context.Context, contact models.Contact, counters mapper.Counters) {
	buf := new(bytes.Buffer)
	hasError := false
	var (
		user           models.AppUser
		warningsCount  int
		transfers      []models.Transfer
		contactBalance models.Balance
	)

	defer func() {
		if r := recover(); r != nil {
			log.Errorf(c, "Panic for Contact(%v): %v", contact.ID, r)
		} else if warningsCount > 0 {
			var logFunc log.LogFunc
			if hasError {
				logFunc = log.Errorf
			} else {
				logFunc = log.Warningf
			}
			userName := ""
			if user.AppUserEntity != nil {
				userName = user.FullName()
			}
			logFunc(c,
				fmt.Sprintf(
					`Contact(id=%v, name=%v): has %v warning, %v transfers
	User(%v): %v
	balance: %v
	`,
					contact.ID,
					contact.FullName(),
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
	transferKeys, err := q.GetAll(c, &transferEntities)
	if err != nil {
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
					return
				} else {
					counterparty.UserName = user.FullName()
				}
			}

			if contact.CounterpartyCounterpartyID !=0 && self.ContactID == 0 {
				self.ContactID = contact.CounterpartyCounterpartyID
				changed = true
			}

			if self.ContactID != 0 && self.ContactName == "" {
				if counterpartyContact, err := dal.Contact.GetContactByID(c, self.ContactID); err != nil {
					log.Errorf(c, err.Error())
					return
				} else {
					self.ContactName = counterpartyContact.FullName()
				}
			}

			if self.UserID != 0 && self.UserName == "" {
				if user, err := dal.User.GetUserByID(c, self.UserID); err != nil {
					log.Errorf(c, err.Error())
					return
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

	logTransfers := func(transfers []models.Transfer, padding int, reset bool) {
		//if reset {
		//	loggedTransfers = make(map[int64]bool, len(transfers))
		//}
		//for _, transfer := range transfers {
		//	logTransfer(transfer, 1)
		//}
	}

	getTransfersBalance := func(transfers []models.Transfer) (totalBalance models.Balance) {
		totalBalance = make(models.Balance)
		for _, transfer := range transfers {
			//logTransfer(transfer, 1)
			switch transfer.DirectionForContact(contact.ID) {
			case models.TransferDirectionUser2Counterparty:
				totalBalance[transfer.Currency] += transfer.AmountInCents
			case models.TransferDirectionCounterparty2User:
				totalBalance[transfer.Currency] -= transfer.AmountInCents
			default:
				panic(fmt.Sprintf("transfer.DirectionForContact(%v): %v", contact.ID, transfer.DirectionForContact(contact.ID)))
			}
		}
		return
	}

	now := time.Now()

	getTransfersOutstanding := func(transfers []models.Transfer) (outstandingBalance models.Balance) {
		outstandingBalance = make(models.Balance)

		for _, transfer := range transfers {
			//logTransfer(transfer, 1)
			switch transfer.DirectionForContact(contact.ID) {
			case models.TransferDirectionUser2Counterparty:
				outstandingBalance[transfer.Currency] += transfer.GetOutstandingValue(now)
			case models.TransferDirectionCounterparty2User:
				outstandingBalance[transfer.Currency] -= transfer.GetOutstandingValue(now)
			default:
				panic(fmt.Sprintf("transfer.DirectionForContact(%v): %v", contact.ID, transfer.DirectionForContact(contact.ID)))
			}
		}
		return
	}

	transfersBalance := getTransfersBalance(transfers)

	verifyReturnIDs := func() (valid bool) {
		valid = true
		for _, transfer := range transfersByID {
			for i, returnTransferID := range transfer.ReturnTransferIDs {
				if _, ok := transfersByID[returnTransferID]; ok {
					m.IncrementCounter(counters, "good_ReturnTransferID", 1)
				} else {
					valid = false
					logTransfer(transfer, 2)
					fmt.Fprintf(buf, "\t\tReturnTransferIDs[%d]: %v\n", i, returnTransferID)
					m.IncrementCounter(counters, "wrong_ReturnTransferID", 1)
					warningsCount += 1
				}
			}
			for i, returnToTransferID := range transfer.ReturnToTransferIDs {
				if _, ok := transfersByID[returnToTransferID]; ok {
					m.IncrementCounter(counters, "good_ReturnToTransferID", 1)
				} else {
					valid = false
					logTransfer(transfer, 2)
					fmt.Fprintf(buf, "\t\tReturnToTransferIDs[%d]: %v\n", i, returnToTransferID)
					m.IncrementCounter(counters, "wrong_ReturnToTransferID", 1)
					warningsCount += 1
				}
			}
		}
		return
	}

	verifyTotals := func() (valid bool) {
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
				m.IncrementCounter(counters, "zero_balance", 1)
				fmt.Fprintf(buf, "\t0 value for currency %v\n", currency)
				warningsCount += 1
			} else {
				m.IncrementCounter(counters, "no_transfers_for_non_zero_balance", 1)
				fmt.Fprintf(buf, "\tno transfers found for %v=%v\n", currency, contactTotal)
				warningsCount += 1
			}
		}
		return
	}

	verifyTotals()

	verifyOutstanding := func(iteration int) (valid bool) {
		valid = true
		transfersOutstanding := getTransfersOutstanding(transfers)
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

	if valid := verifyOutstanding(1); !valid {
		//rollingBalance := make(models.Balance, len(transfersBalance)+1)

		fmt.Fprintf(buf, "Will try to fix %d transfers:\n", len(transfers))
		logTransfers(transfers, 1, true)

		loggedTransfers = make(map[int64]bool, len(transfers))
		transfersByCurrency := make(map[models.Currency][]models.Transfer)

		transfersToSave := make(map[int64]*models.TransferEntity)

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

		for currency, currencyTransfers := range transfersByCurrency {
			fmt.Fprintf(buf, "\tcurrency: %v - %d transfers\n", currency, len(currencyTransfers))
		}

		if valid := verifyOutstanding(2); !valid {
			fmt.Fprint(buf, "Outstandings are invalid after fix")
		} else if valid = verifyTotals(); !valid {
			fmt.Fprint(buf, "Totals are invalid after fix")
		} else if valid = verifyReturnIDs(); !valid {
			fmt.Fprint(buf, "Return IDs are invalid after fix")
		} else {
			fmt.Fprintf(buf, "%v transfers to save!\n", len(transfersToSave))
			logTransfers(transfers, 1, true)
			entitiesToSave := make([]db.EntityHolder, 0, len(transfersToSave))
			for id, transfer := range transfersToSave {
				entitiesToSave = append(entitiesToSave, &models.Transfer{ID: id, TransferEntity: transfer})
			}
			//gaedb.LoggingEnabled = true
			//if err = dal.DB.UpdateMulti(c, entitiesToSave); err != nil {
			//	gaedb.LoggingEnabled = false
			//	fmt.Fprintf(buf, "ERROR: failed to save transfers: "+err.Error())
			//	hasError = true
			//	return
			//}
			//gaedb.LoggingEnabled = false
			//fmt.Fprintf(buf, "SAVED %v transfers!\n", len(entitiesToSave))
		}
	}

	if warningsCount == 0 {
		m.IncrementCounter(counters, "good_contacts", 1)
		//log.Infof(c, contactPrefix + "is OK, %v transfers", len(transfers))
	} else {
		m.Lock()
		counters.Increment("bad_contacts", 1)
		counters.Increment("warnings", int64(warningsCount))
		m.Unlock()

		if user, err = dal.User.GetUserByID(c, contact.UserID); err != nil {
			log.Errorf(c, errors.WithMessage(err, fmt.Sprintf("Contact(%v): ", contact.ID)+"user not found by ID").Error())
			return
		}

		contactBalance := contact.Balance()

		if len(contactBalance) == 0 {
			contactBalance = nil
		}
	}
}
