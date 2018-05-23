package inspector

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type contactWithBalances struct {
	models.Contact
	transfersCount int
	balances       balances
}

type transfersInfo struct {
	count   int
	balance models.Balance
}

func newContactWithBalances(now time.Time, contact models.Contact) contactWithBalances {
	balanceWithInterest, err := contact.BalanceWithInterest(nil, now)
	result := contactWithBalances{
		Contact:  contact,
		balances: newBalances("contact", contact.Balance(), balanceWithInterest),
	}
	result.balances.withInterest.err = err
	return result
}

func newBalanceSummary(who string, balance models.Balance) (balances balancesByCurrency) {
	balances = balancesByCurrency{
		Mutex:      new(sync.Mutex),
		byCurrency: make(map[models.Currency]balanceRow, len(balance)),
	}
	for currency, value := range balance {
		row := balances.byCurrency[currency]
		switch who {
		case "user":
			row.user = value
		case "contact":
			row.contacts = value
		default:
			panic("unknown who: " + who)
		}

		balances.byCurrency[currency] = row
	}
	return
}

func (bs balancesByCurrency) SetBalance(setter func(bs balancesByCurrency)) {
	bs.Lock()
	setter(bs)
	bs.Unlock()
}

func validateTransfers(c context.Context, userID int64, userBalances balances) (
	byContactWithoutInterest map[int64]transfersInfo, err error,
) {
	query := datastore.NewQuery(models.TransferKind).Filter("BothUserIDs=", userID)

	byContactWithoutInterest = make(map[int64]transfersInfo)

	iterator := query.Run(c)

	for {
		transferEntity := new(models.TransferEntity)
		if _, err = iterator.Next(transferEntity); err != nil {
			if err == datastore.Done {
				break
			}
			panic(err)
		}
		userBalances.withoutInterest.Lock()
		row := userBalances.withoutInterest.byCurrency[transferEntity.Currency]
		contactID := transferEntity.To().ContactID
		var direction decimal.Decimal64p2
		switch {

		}
		switch userID {
		case transferEntity.From().UserID:
			direction = 1
		case transferEntity.To().UserID:
			direction = -1
		default:
			direction = 0
		}
		row.transfers += direction * transferEntity.AmountInCents
		if contactTransfersInfo, ok := byContactWithoutInterest[contactID]; ok {
			contactTransfersInfo.count += 1
			contactTransfersInfo.balance[transferEntity.Currency] += direction * transferEntity.AmountInCents
			byContactWithoutInterest[contactID] = contactTransfersInfo
		} else {
			byContactWithoutInterest[contactID] = transfersInfo{
				count:   1,
				balance: models.Balance{transferEntity.Currency: direction * transferEntity.AmountInCents},
			}
		}
		userBalances.withoutInterest.byCurrency[transferEntity.Currency] = row
		userBalances.withoutInterest.Unlock()
	}
	return
}

func validateContacts(c context.Context,
	now time.Time,
	user models.AppUser,
	userBalances balances,
) (
	contactsMissingInJson, contactsMissedByQuery, matchedContacts []contactWithBalances,
	contactInfosNotFoundInDb []models.UserContactJson,
	err error,
) {
	userContactsJson := user.Contacts()
	contactInfos := make([]contactWithBalances, len(userContactsJson))
	contactInfosByID := make(map[int64]contactWithBalances, len(contactInfos))

	contactsTotalWithoutInterest := make(models.Balance, len(userBalances.withoutInterest.byCurrency))
	contactsTotalWithInterest := make(models.Balance, len(userBalances.withInterest.byCurrency))

	updateBalance := func(contact models.Contact) (ci contactWithBalances) {
		contactBalanceWithoutInterest := contact.Balance()
		contactBalanceWithInterest, err := contact.BalanceWithInterest(c, now)
		if err == nil {

		}
		ci = newContactWithBalances(now, contact)
		for currency, value := range contactBalanceWithoutInterest {
			contactsTotalWithoutInterest[currency] += value
		}
		for currency, value := range contactBalanceWithInterest {
			contactsTotalWithInterest[currency] += value
		}
		for _, userContactJson := range userContactsJson {
			if userContactJson.ID == contact.ID {
				for currency, value := range userContactJson.Balance() {
					row := ci.balances.withoutInterest.byCurrency[currency]
					row.user = value
					ci.balances.withoutInterest.byCurrency[currency] = row
				}

				if userContactBalanceWithInterest, err := userContactJson.BalanceWithInterest(nil, now); err != nil {
					ci.balances.withInterest.err = err
				} else {
					for currency, value := range userContactBalanceWithInterest {
						row := ci.balances.withInterest.byCurrency[currency]
						row.user = value
						ci.balances.withInterest.byCurrency[currency] = row
					}
				}
				break
			}
		}
		return
	}

	for i, userContactInfo := range userContactsJson {
		var contact models.Contact
		if contact, err = facade.GetContactByID(c, userContactInfo.ID); err != nil {
			if db.IsNotFound(err) {
				contactInfosNotFoundInDb = append(contactInfosNotFoundInDb, userContactInfo)
			} else {
				panic(err)
			}
		}
		contactInfos[i] = updateBalance(contact)
		contactInfosByID[contact.ID] = contactInfos[i]
	}

	for _, contact := range contactInfos {
		for _, userContact := range userContactsJson {
			if userContact.ID == contact.ID {
				goto foundInUserJson
			}
		}
		contactsMissingInJson = append(contactsMissingInJson, contact)
	foundInUserJson:
	}

	query := datastore.NewQuery(models.ContactKind).Filter("UserID=", user.ID).KeysOnly()

	iterator := query.Run(c)

	for {
		var key *datastore.Key
		if key, err = iterator.Next(nil); err != nil {
			if err == datastore.Done {
				break
			}
			panic(err)
		}
		if contactInfo, ok := contactInfosByID[key.IntID()]; ok {
			matchedContacts = append(matchedContacts, contactInfo)
		} else {
			var contact models.Contact
			if contact, err = facade.GetContactByID(c, key.IntID()); err != nil {
				return
			}
			contactInfo = updateBalance(contact)
			contactInfos = append(contactInfos, contactInfo)
			contactsMissingInJson = append(contactInfos, contactInfo)
		}
	}

	defer func() {
		log.Debugf(c, "contactInfos: %v", contactInfos)
		log.Debugf(c, "contactsMissingInJson: %v", contactsMissingInJson)
		log.Debugf(c, "contactsMissedByQuery: %v", contactsMissedByQuery)
		log.Debugf(c, "matchedContacts: %v", matchedContacts)
	}()

	log.Debugf(c, "contactsTotalWithoutInterest: %v", contactsTotalWithoutInterest)
	log.Debugf(c, "contactsTotalWithInterest: %v", contactsTotalWithInterest)

	userBalances.withoutInterest.SetBalance(func(balances balancesByCurrency) {
		for currency, value := range contactsTotalWithoutInterest {
			row := balances.byCurrency[currency]
			row.contacts += value
			balances.byCurrency[currency] = row
		}
	})
	userBalances.withInterest.SetBalance(func(balances balancesByCurrency) {
		for currency, value := range contactsTotalWithInterest {
			row := balances.byCurrency[currency]
			row.contacts += value
			balances.byCurrency[currency] = row
		}
	})
	return
}

func userPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	userID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var (
		user                                                          models.AppUser
		contactsMissingInJson, contactsMissedByQuery, matchedContacts []contactWithBalances
		contactInfosNotFoundInDb                                      []models.UserContactJson
	)
	if user, err = facade.User.GetUserByID(c, userID); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	wg := new(sync.WaitGroup)

	now := time.Now()

	userBalanceWithInterest, err := user.BalanceWithInterest(c, now)
	userBalances := newBalances("user", user.Balance(), userBalanceWithInterest)
	if err != nil {
		userBalances.withInterest.err = err
	}

	wg.Add(1)
	go func() { // TODO: Move to DAL?
		defer wg.Done()
		contactsMissingInJson, contactsMissedByQuery, matchedContacts, contactInfosNotFoundInDb, err = validateContacts(c, now, user, userBalances)
	}()

	var byContactWithoutInterest map[int64]transfersInfo
	wg.Add(1)
	go func() {
		defer wg.Done()
		byContactWithoutInterest, err = validateTransfers(c, userID, userBalances)
	}()

	wg.Wait()

	for contactID, contactTransfersInfo := range byContactWithoutInterest {
		for i, contactInfo := range matchedContacts {
			if contactInfo.ID == contactID {
				contactInfo.transfersCount = contactTransfersInfo.count
				for currency, value := range contactTransfersInfo.balance {
					row := contactInfo.balances.withoutInterest.byCurrency[currency]
					row.transfers = value
					contactInfo.balances.withoutInterest.byCurrency[currency] = row
				}
				matchedContacts[i] = contactInfo
				break
			}
		}
	}

	log.Debugf(c, "matchedContacts: %v", matchedContacts)

	renderUserPage(now,
		user,
		userBalances,
		contactsMissingInJson, contactsMissedByQuery, matchedContacts, contactInfosNotFoundInDb,
		w)
}
