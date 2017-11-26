package support

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal/gaedal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"github.com/strongo/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"
)

func InitSupportHandlers(router *httprouter.Router) {
	router.HandlerFunc("GET", "/support/validate-users", ValidateUsersHandler)
	router.HandlerFunc("GET", "/support/validate-user", ValidateUserHandler)
}

func ValidateUsersHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	fix := r.URL.Query().Get("fix")
	query := datastore.NewQuery(models.AppUserKind).KeysOnly() //.Limit(25)
	t := query.Run(c)
	batchSize := 100
	tasks := make([]*taskqueue.Task, 0, batchSize)
	var (
		usersCount int
		params     url.Values
	)

	addTasksToQueue := func() error {
		if _, err := taskqueue.AddMulti(c, tasks, "support"); err != nil {
			log.Errorf(c, "Failed to add tasks: %v", err)
			return err
		}
		tasks = make([]*taskqueue.Task, 0, batchSize)
		return nil
	}

	for {
		if key, err := t.Next(nil); err != nil {
			if err == datastore.Done {
				break
			}
			log.Errorf(c, "Failed to fetch %v: %v", key, err)
			return
		} else {
			usersCount += 1
			taskUrl := fmt.Sprintf("/support/validate-user?id=%v", key.IntID())
			if fix != "" {
				taskUrl += "&fix=" + fix
			}
			tasks = append(tasks, taskqueue.NewPOSTTask(taskUrl, params))
			if len(tasks) == batchSize {
				if err = addTasksToQueue(); err != nil {
					return
				}
			}
		}

	}
	if len(tasks) > 0 {
		if err := addTasksToQueue(); err != nil {
			return
		}
	}
	log.Errorf(c, "(NOT error) Users count: %v", usersCount)
	w.Write([]byte(fmt.Sprintf("Users count: %v", usersCount)))
}

type int64sortable []int64

func (a int64sortable) Len() int           { return len(a) }
func (a int64sortable) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64sortable) Less(i, j int) bool { return a[i] < a[j] }

func ValidateUserHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	doFixes := r.URL.Query().Get("fix") == "all"
	userID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		log.Errorf(c, "Failed to parse 'id' parameter: %v", err)
		return
	}
	userKey := gaedal.NewAppUserKey(c, userID)
	var user models.AppUserEntity
	if err = nds.Get(c, userKey, &user); err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Errorf(c, "User not found by key=%v", userKey)
		} else {
			log.Errorf(c, "Failed to get user by key=%v: %v", userKey, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	query := datastore.NewQuery(models.ContactKind).Filter("UserID =", userID)
	var userCounterparties []*models.ContactEntity
	counterpartyKeys, err := query.GetAll(c, &userCounterparties)
	if err != nil {
		log.Errorf(c, "Failed to load user counterparties: %v", err)
		return
	}

	userCounterpartyIDs := make(int64sortable, len(user.ContactIDs()))
	for i, v := range user.ContactIDs() {
		userCounterpartyIDs[i] = v
	}

	if user.TotalContactsCount() != len(userCounterpartyIDs) {
		log.Warningf(c, "user.TotalContactsCount() != len(user.ContactIDs()) => %v != %v", user.TotalContactsCount(), len(userCounterpartyIDs))
	}

	sort.Sort(userCounterpartyIDs)

	counterpartyIDs := make(int64sortable, len(counterpartyKeys))
	for i, v := range counterpartyKeys {
		counterpartyIDs[i] = v.IntID()
	}
	sort.Sort(counterpartyIDs)

	query = datastore.NewQuery(models.TransferKind).Filter("BothUserIDs =", userID).Order("DtCreated")

	var transferEntities []*models.TransferEntity
	transferKeys, err := query.GetAll(c, &transferEntities)
	if err != nil {
		log.Errorf(c, "Failed to load transfers: %v", err)
		return
	}

	type transfersInfo struct {
		Count  int
		LastID int64
		LastAt time.Time
	}

	transfersInfoByCounterparty := make(map[int64]transfersInfo, len(counterpartyIDs))

	for i, transferEntity := range transferEntities {
		counterpartyInfo := transferEntity.CounterpartyInfoByUserID(userID)
		counterpartyTransfersInfo := transfersInfoByCounterparty[counterpartyInfo.ContactID]
		counterpartyTransfersInfo.Count += 1
		if counterpartyTransfersInfo.LastAt.Before(transferEntity.DtCreated) {
			counterpartyTransfersInfo.LastAt = transferEntity.DtCreated
			counterpartyTransfersInfo.LastID = transferKeys[i].IntID()
		}
		transfersInfoByCounterparty[counterpartyInfo.ContactID] = counterpartyTransfersInfo
	}

	fixUserCounterparties := func() {
		var txUser models.AppUserEntity
		err := dal.DB.RunInTransaction(c, func(c context.Context) error {
			log.Debugf(c, "Transaction started..")
			if err := nds.Get(c, userKey, &txUser); err != nil {
				return errors.Wrap(err, "Failed to get user by key")
			}
			if txUser.SavedCounter != user.SavedCounter {
				return fmt.Errorf("User changed since last load: txUser.SavedCounter:%v != user.SavedCounter:%v", txUser.SavedCounter, user.SavedCounter)
			}
			txUser.ContactsJson = ""
			for i, counterpartyEntity := range userCounterparties {
				counterpartyID := counterpartyKeys[i].IntID()
				if counterpartyTransfersInfo, ok := transfersInfoByCounterparty[counterpartyID]; ok {
					counterpartyEntity.LastTransferAt = counterpartyTransfersInfo.LastAt
					counterpartyEntity.LastTransferID = counterpartyTransfersInfo.LastID
					counterpartyEntity.CountOfTransfers = counterpartyTransfersInfo.Count
				} else {
					counterpartyEntity.CountOfTransfers = 0
					counterpartyEntity.LastTransferAt = time.Time{}
					counterpartyEntity.LastTransferID = 0
				}
				models.AppUser{IntegerID: db.NewIntID(userID), AppUserEntity: &txUser}.AddOrUpdateContact(models.NewContact(counterpartyID, counterpartyEntity))
			}
			if _, err = nds.Put(c, userKey, &txUser); err != nil {
				return errors.Wrap(err, "Failed to save fixed user")
			}
			return nil
		}, nil)
		if err != nil {
			log.Errorf(c, "Failed to fix user.CounterpartyIDs: %v", err)
			return
		}
		log.Infof(c, "Fixed user.ContactsJson\n\tfrom: %v\n\tto: %v", user.ContactsJson, txUser.ContactsJson)
		user = txUser
	}

	if len(userCounterpartyIDs) != len(counterpartyIDs) {
		log.Warningf(c, "len(userCounterpartyIDs) != len(counterpartyIDs) => %v != %v", len(userCounterpartyIDs), len(counterpartyIDs))
		if doFixes {
			fixUserCounterparties()
		} else {
			return // Do not continue if counterparties are not in sync
		}
	} else {
		for i, v := range userCounterpartyIDs {
			if counterpartyIDs[i] != v {
				log.Warningf(c, "user.CounterpartyIDs != counterpartyKeys\n\tuserCounterpartyIDs:\n\t\t%v\n\tcounterpartyIDs:\n\t\t%v", userCounterpartyIDs, counterpartyIDs)
				if doFixes {
					fixUserCounterparties()
					break
				} else {
					return // Do not continue if counterparties are not in sync
				}
			}
		}
	}
	log.Infof(c, "OK - User ContactsJson is OK")

	// We need counterparties by ID to check balance against transfers
	counterpartiesByID := make(map[int64]*models.ContactEntity, len(counterpartyIDs))
	for i, counterpartyKey := range counterpartyKeys {
		counterpartiesByID[counterpartyKey.IntID()] = userCounterparties[i]
	}

	if len(transferKeys) > 0 && user.LastTransferID == 0 {
		if doFixes {
			var txUser models.AppUserEntity
			err = dal.DB.RunInTransaction(c, func(c context.Context) error {
				if err := nds.Get(c, userKey, &txUser); err != nil {
					return err
				}
				if txUser.LastTransferID == 0 {
					i := len(transferKeys) - 1
					txUser.LastTransferID = transferKeys[i].IntID()
					txUser.LastTransferAt = transferEntities[i].DtCreated
					_, err = nds.Put(c, userKey, &txUser)
					return err
				}
				return nil
			}, nil)
			if err != nil {
				log.Errorf(c, "Failed to update user.LastTransferID")
			} else {
				log.Infof(c, "Fixed user.LastTransferID")
				user = txUser
			}
		} else {
			log.Warningf(c, "user.LastTransferID is not set")
		}
	}

	// Get stored user total balance
	userBalance := user.Balance()

	transfersBalanceByCounterpartyID := make(map[int64]models.Balance, len(counterpartyIDs))

	for i, transfer := range transferEntities {
		var counterpartyID int64
		switch userID {
		case transfer.CreatorUserID:
			counterpartyID = transfer.Counterparty().ContactID
		case transfer.Counterparty().UserID:
			counterpartyID = transfer.Creator().ContactID
		default:
			log.Errorf(c, "userID=%v is NOT equal to transfer.CreatorUserID=%v or transfer.Contact().UserID=%v", userID, transfer.CreatorUserID, transfer.Counterparty().UserID)
			return
		}
		transfersCounterpartyBalance, ok := transfersBalanceByCounterpartyID[counterpartyID]
		if !ok {
			transfersCounterpartyBalance = make(models.Balance)
			transfersBalanceByCounterpartyID[counterpartyID] = transfersCounterpartyBalance
		}
		value := transfer.GetAmount().Value
		currency := models.Currency(transfer.Currency)
		switch transfer.DirectionForUser(userID) {
		case models.TransferDirectionUser2Counterparty:
			transfersCounterpartyBalance[currency] += value
		case models.TransferDirectionCounterparty2User:
			transfersCounterpartyBalance[currency] -= value
		default:
			log.Errorf(c, "Transfer %v has unknown direction: %v", transferKeys[i], transfer.DirectionForUser(userID))
			return
		}
	}

	//log.Debugf(c, "transfersBalanceByCounterpartyID: %v", transfersBalanceByCounterpartyID)

	transfersTotalBalance := make(models.Balance)
	for _, transfersCounterpartyBalance := range transfersBalanceByCounterpartyID {
		for currency, value := range transfersCounterpartyBalance {
			if value == 0 {
				delete(transfersCounterpartyBalance, currency)
			} else {
				transfersTotalBalance[currency] += value
			}
		}
	}

	for currency, value := range transfersTotalBalance {
		if value == 0 {
			delete(transfersTotalBalance, currency)
		}
	}

	if len(userBalance) != len(transfersTotalBalance) {
		log.Warningf(c, "len(userBalance) != len(transfersTotalBalance) =>\n\t%d: %v\n\t%d: %v", len(userBalance), userBalance, len(transfersTotalBalance), transfersTotalBalance)
	}

	userBalanceIsOK := true

	for currency, userVal := range userBalance {
		if transfersVal, ok := transfersTotalBalance[currency]; !ok {
			log.Warningf(c, "User has %v=%v balance but no corresponding transfers' balance.", currency, userVal)
			userBalanceIsOK = false
		} else if transfersVal != userVal {
			log.Warningf(c, "Currency(%v) User balance %v not equal to transfers' balance %v", currency, userVal, transfersVal)
			userBalanceIsOK = false
		}
	}

	for currency, transfersVal := range transfersTotalBalance {
		if _, ok := userBalance[currency]; !ok {
			log.Warningf(c, "Transfers has %v=%v balance but no corresponding user balance.", currency, transfersVal)
			userBalanceIsOK = false
		}
	}

	if userBalanceIsOK {
		log.Infof(c, "OK - User.Balance() is matching to %v transfers' balance.", len(transferEntities))
	} else {
		log.Warningf(c, "Calculated balance for %v user transfers does not match user's total balance.", len(transferEntities))
		if !doFixes {
			log.Debugf(c, "Pass fix=all to fix user balance")
		} else {
			err = dal.DB.RunInTransaction(c, func(c context.Context) error {
				var txUser models.AppUserEntity
				if err := nds.Get(c, userKey, &txUser); err != nil {
					return errors.Wrapf(err, "Failed to get by key=%v", userKey)
				}
				if !reflect.DeepEqual(txUser.BalanceJson, user.BalanceJson) {
					return errors.New("User changed: !reflect.DeepEqual(txUser.Balance(), user.Balance())")
				}

				if balanceJson, err := json.Marshal(transfersTotalBalance); err != nil {
					return err
				} else {
					txUser.BalanceJson = string(balanceJson)
					txUser.BalanceCount = len(transfersTotalBalance)
					if _, err = nds.Put(c, userKey, &txUser); err != nil {
						return errors.Wrap(err, "Failed to save user with fixed balance")
					}
				}
				return nil
			}, nil)
			if err != nil {
				err = errors.Wrap(err, "Failed to fix user balance")
				log.Errorf(c, err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			log.Infof(c, "Fixed user balance")
		}
	}

	var counterpartyIDsWithMatchingBalance, counterpartyIDsWithNonMatchingBalance []int64

	for i, counterpartyKey := range counterpartyKeys {
		counterpartyID := counterpartyKey.IntID()
		counterparty := userCounterparties[i]
		counterpartyBalance := counterparty.Balance()

		if transfersCounterpartyBalance := transfersBalanceByCounterpartyID[counterpartyID]; (len(transfersCounterpartyBalance) == 0 && len(counterpartyBalance) == 0) || reflect.DeepEqual(transfersCounterpartyBalance, counterpartyBalance) {
			counterpartyIDsWithMatchingBalance = append(counterpartyIDsWithMatchingBalance, counterpartyID)
			if counterparty.BalanceCount != len(counterpartyBalance) {
				if doFixes {
					var txCounterparty models.ContactEntity
					err = dal.DB.RunInTransaction(c, func(c context.Context) error {
						if err := nds.Get(c, counterpartyKey, &txCounterparty); err != nil {
							return err
						}
						counterpartyBalance := txCounterparty.Balance()
						balanceCount := len(counterpartyBalance)
						if txCounterparty.BalanceCount != balanceCount {
							txCounterparty.BalanceCount = balanceCount
							_, err = nds.Put(c, counterpartyKey, &txCounterparty)
							return err
						}
						return nil
					}, nil)
					if err != nil {
						log.Errorf(c, "Failed to fix counterparty.BalanceCount, ID=%v", counterpartyID)
					} else {
						log.Warningf(c, "Fixed counterparrty.BalanceCount, ID=%v", counterpartyID)
					}
				} else {
					log.Warningf(c, "counterparty.BalanceCount != len(counterparty.BalanceCount), ID: %v", counterpartyID)
				}
			}
		} else {
			counterpartyIDsWithNonMatchingBalance = append(counterpartyIDsWithNonMatchingBalance, counterpartyID)
			log.Warningf(c, "Contact ID=%v has balance not matching transfers' balance:\n\tContact: %v\n\tTransfers: %v", counterpartyID, counterpartyBalance, transfersCounterpartyBalance)
			if doFixes {
				var txCounterparty models.ContactEntity
				err := dal.DB.RunInTransaction(c, func(c context.Context) error {
					if err := nds.Get(c, counterpartyKey, &txCounterparty); err != nil {
						return errors.Wrapf(err, "Failed to get by key=%v", counterpartyKey)
					}
					if !reflect.DeepEqual(txCounterparty.BalanceJson, counterparty.BalanceJson) {
						return errors.New("Contact changed since check: !reflect.DeepEqual(txCounterparty.Balance(), counterparty.Balance())")
					}

					if balanceJson, err := json.Marshal(transfersCounterpartyBalance); err != nil {
						return errors.Wrap(err, "Failed to json.Marshal(transfersCounterpartyBalance)")
					} else {
						txCounterparty.BalanceJson = string(balanceJson)
						txCounterparty.BalanceCount = len(transfersCounterpartyBalance)
						if _, err := nds.Put(c, counterpartyKey, &txCounterparty); err != nil {
							return errors.Wrapf(err, "Failed to save counterparty with ID=%v", counterpartyID)
						}
					}
					return nil
				}, nil)
				if err != nil {
					log.Errorf(c, "Failed to fix counterparty with ID=%v: %v", counterpartyID, err)
				} else {
					log.Infof(c, "Fixed counterparty with ID=%v", counterpartyID)
					userCounterparties[i] = &txCounterparty
				}
			}
		}
	}
	if len(counterpartyIDsWithMatchingBalance) > 0 {
		log.Infof(c, "There are %v counterparties with balance matching to transfers: %v", len(counterpartyIDsWithMatchingBalance), counterpartyIDsWithMatchingBalance)
	}
	if len(counterpartyIDsWithNonMatchingBalance) > 0 {
		log.Warningf(c, "There are %v counterparties with balance NOT matching to transfers: %v", len(counterpartyIDsWithNonMatchingBalance), counterpartyIDsWithNonMatchingBalance)
	}
}
