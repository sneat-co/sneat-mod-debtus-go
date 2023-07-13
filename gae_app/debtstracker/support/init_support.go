package support

import (
	"encoding/json"
	"fmt"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"google.golang.org/appengine/v2"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"time"

	"context"
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
	"google.golang.org/appengine/v2/taskqueue"
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
	_, _ = w.Write([]byte(fmt.Sprintf("Users count: %v", usersCount)))
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
	user := models.NewAppUser(userID, nil)
	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		log.Errorf(c, "Failed to get database: %v", err)
		return
	}
	if err = db.Get(c, user.Record); err != nil {
		if dal.IsNotFound(err) {
			log.Errorf(c, "User not found by key: %v", err)
		} else {
			log.Errorf(c, "Failed to get user by key=%v: %v", user.Key, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	query := dal.From(models.ContactKind).WhereField("UserID", dal.Equal, userID).SelectInto(func() dal.Record {
		return dal.NewRecordWithoutKey(models.AppUserKind, reflect.Int64, new(models.AppUserData))
	})
	userCounterpartyRecords, err := db.QueryAllRecords(c, query)
	if err != nil {
		log.Errorf(c, "Failed to load user counterparties: %v", err)
		return
	}

	userCounterpartyIDs := make(int64sortable, len(user.Data.ContactIDs()))
	for i, v := range user.Data.ContactIDs() {
		userCounterpartyIDs[i] = v
	}

	if user.Data.TotalContactsCount() != len(userCounterpartyIDs) {
		log.Warningf(c, "user.TotalContactsCount() != len(user.ContactIDs()) => %v != %v", user.Data.TotalContactsCount(), len(userCounterpartyIDs))
	}

	sort.Sort(userCounterpartyIDs)

	counterpartyIDs := make(int64sortable, len(userCounterpartyRecords))
	for i, v := range userCounterpartyRecords {
		counterpartyIDs[i] = v.Key().ID.(int64)
	}
	sort.Sort(counterpartyIDs)

	query = dal.From(models.TransferKind).WhereField("BothUserIDs", dal.Equal, userID).OrderBy(dal.AscendingField("DtCreated")).SelectInto(func() dal.Record {
		return dal.NewRecordWithoutKey(models.AppUserKind, reflect.Int64, new(models.AppUserData))
	})

	transferRecords, err := db.QueryAllRecords(c, query)

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

	for _, transferRecord := range transferRecords {
		transferEntity := transferRecord.Data().(*models.TransferData)
		counterpartyInfo := transferEntity.CounterpartyInfoByUserID(userID)
		counterpartyTransfersInfo := transfersInfoByCounterparty[counterpartyInfo.ContactID]
		counterpartyTransfersInfo.Count += 1
		if counterpartyTransfersInfo.LastAt.Before(transferEntity.DtCreated) {
			counterpartyTransfersInfo.LastAt = transferEntity.DtCreated
			counterpartyTransfersInfo.LastID = transferRecord.Key().ID.(int64)
		}
		transfersInfoByCounterparty[counterpartyInfo.ContactID] = counterpartyTransfersInfo
	}

	fixUserCounterparties := func() {
		var txUser models.AppUser
		var db dal.Database
		if db, err = facade.GetDatabase(c); err != nil {
			log.Errorf(c, "Failed to get database: %v", err)
			return
		}
		err := db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
			log.Debugf(c, "Transaction started..")
			txUser = models.NewAppUser(userID, nil)
			if err := tx.Get(c, txUser.Record); err != nil {
				return err
			}
			if txUser.Data.SavedCounter != user.Data.SavedCounter {
				return fmt.Errorf("user changed since last load: txUser.SavedCounter:%v != user.SavedCounter:%v", txUser.Data.SavedCounter, user.Data.SavedCounter)
			}
			txUser.Data.ContactsJson = ""
			for _, counterpartyRecord := range userCounterpartyRecords {
				counterpartyEntity := counterpartyRecord.Data().(*models.ContactData)
				counterpartyID := counterpartyRecord.Key().ID.(int64)
				if counterpartyTransfersInfo, ok := transfersInfoByCounterparty[counterpartyID]; ok {
					counterpartyEntity.LastTransferAt = counterpartyTransfersInfo.LastAt
					counterpartyEntity.LastTransferID = counterpartyTransfersInfo.LastID
					counterpartyEntity.CountOfTransfers = counterpartyTransfersInfo.Count
				} else {
					counterpartyEntity.CountOfTransfers = 0
					counterpartyEntity.LastTransferAt = time.Time{}
					counterpartyEntity.LastTransferID = 0
				}
				txUser.AddOrUpdateContact(models.NewContact(counterpartyID, counterpartyEntity))
			}
			if err = tx.Set(c, txUser.Record); err != nil {
				return fmt.Errorf("failed to save fixed user: %w", err)
			}
			return nil
		}, nil)
		if err != nil {
			log.Errorf(c, "Failed to fix user.CounterpartyIDs: %v", err)
			return
		}
		log.Infof(c, "Fixed user.ContactsJson\n\tfrom: %v\n\tto: %v", user.Data.ContactsJson, txUser.Data.ContactsJson)
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
	counterpartiesByID := make(map[int64]*models.ContactData, len(counterpartyIDs))
	for _, counterpartyRecord := range userCounterpartyRecords {
		counterpartiesByID[counterpartyRecord.Key().ID.(int64)] = counterpartyRecord.Data().(*models.ContactData)
	}

	if len(transferRecords) > 0 && user.Data.LastTransferID == 0 {
		if doFixes {
			var txUser models.AppUser
			var db dal.Database
			if db, err = facade.GetDatabase(c); err != nil {
				log.Errorf(c, "Failed to get database: %v", err)
				return
			}
			err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
				txUser = models.NewAppUser(userID, nil)
				if err = tx.Get(c, txUser.Record); err != nil {
					return err
				}
				if txUser.Data.LastTransferID == 0 {
					i := len(transferRecords) - 1
					txUser.Data.LastTransferID = transferRecords[i].Key().ID.(int64)
					txUser.Data.LastTransferAt = transferRecords[i].Data().(*models.TransferData).DtCreated
					err = tx.Set(c, txUser.Record)
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
	userBalance := user.Data.Balance()

	transfersBalanceByCounterpartyID := make(map[int64]money.Balance, len(counterpartyIDs))

	for i, transferRecord := range transferRecords {
		transferData := transferRecord.Data().(*models.TransferData)
		var counterpartyID int64
		switch userID {
		case transferData.CreatorUserID:
			counterpartyID = transferData.Counterparty().ContactID
		case transferData.Counterparty().UserID:
			counterpartyID = transferData.Creator().ContactID
		default:
			log.Errorf(c, "userID=%v is NOT equal to transferData.CreatorUserID=%v or transferData.Contact().UserID=%v", userID, transferData.CreatorUserID, transferData.Counterparty().UserID)
			return
		}
		transfersCounterpartyBalance, ok := transfersBalanceByCounterpartyID[counterpartyID]
		if !ok {
			transfersCounterpartyBalance = make(money.Balance)
			transfersBalanceByCounterpartyID[counterpartyID] = transfersCounterpartyBalance
		}
		value := transferData.GetAmount().Value
		currency := money.Currency(transferData.Currency)
		switch transferData.DirectionForUser(userID) {
		case models.TransferDirectionUser2Counterparty:
			transfersCounterpartyBalance[currency] += value
		case models.TransferDirectionCounterparty2User:
			transfersCounterpartyBalance[currency] -= value
		default:
			log.Errorf(c, "Transfer %v has unknown direction: %v", transferRecords[i].Key().ID, transferData.DirectionForUser(userID))
			return
		}
	}

	//log.Debugf(c, "transfersBalanceByCounterpartyID: %v", transfersBalanceByCounterpartyID)

	transfersTotalBalance := make(money.Balance)
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
		log.Infof(c, "OK - User.Balance() is matching to %v transfers' balance.", len(transferRecords))
	} else {
		log.Warningf(c, "Calculated balance for %v user transfers does not match user's total balance.", len(transferRecords))
		if !doFixes {
			log.Debugf(c, "Pass fix=all to fix user balance")
		} else {
			var db dal.Database
			if db, err = facade.GetDatabase(c); err != nil {
				log.Errorf(c, "Failed to get database: %v", err)
				return
			}
			err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
				txUser := models.NewAppUser(userID, nil)
				if err := tx.Get(c, txUser.Record); err != nil {
					return err
				}
				if !reflect.DeepEqual(txUser.Data.BalanceJson, user.Data.BalanceJson) {
					return errors.New("user changed: !reflect.DeepEqual(txUser.Balance(), user.Balance())")
				}

				if balanceJson, err := json.Marshal(transfersTotalBalance); err != nil {
					return err
				} else {
					txUser.Data.BalanceJson = string(balanceJson)
					txUser.Data.BalanceCount = len(transfersTotalBalance)
					if err = tx.Set(c, txUser.Record); err != nil {
						return fmt.Errorf("failed to save user with fixed balance: %w", err)
					}
				}
				return nil
			}, nil)
			if err != nil {
				err = fmt.Errorf("failed to fix user balance: %w", err)
				log.Errorf(c, err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			log.Infof(c, "Fixed user balance")
		}
	}

	var counterpartyIDsWithMatchingBalance, counterpartyIDsWithNonMatchingBalance []int64

	for _, counterpartyRecord := range userCounterpartyRecords {
		counterpartyKey := counterpartyRecord.Key()
		counterpartyID := counterpartyKey.ID.(int64)
		counterparty := counterpartyRecord.Data().(*models.ContactData)
		counterpartyBalance := counterparty.Balance()

		if transfersCounterpartyBalance := transfersBalanceByCounterpartyID[counterpartyID]; (len(transfersCounterpartyBalance) == 0 && len(counterpartyBalance) == 0) || reflect.DeepEqual(transfersCounterpartyBalance, counterpartyBalance) {
			counterpartyIDsWithMatchingBalance = append(counterpartyIDsWithMatchingBalance, counterpartyID)
			if counterparty.BalanceCount != len(counterpartyBalance) {
				if doFixes {
					var db dal.Database
					if db, err = facade.GetDatabase(c); err != nil {
						log.Errorf(c, "Failed to get database: %v", err)
						return
					}
					err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
						txCounterparty := models.NewContact(counterpartyKey.ID.(int64), nil)
						if err = tx.Get(c, txCounterparty.Record); err != nil {
							return err
						}
						counterpartyBalance := txCounterparty.Data.Balance()
						balanceCount := len(counterpartyBalance)
						if txCounterparty.Data.BalanceCount != balanceCount {
							txCounterparty.Data.BalanceCount = balanceCount
							return tx.Set(c, txCounterparty.Record)
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
				//var txCounterparty models.Contact
				var db dal.Database
				if db, err = facade.GetDatabase(c); err != nil {
					log.Errorf(c, "Failed to get database: %v", err)
					return
				}
				err := db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
					txCounterparty := models.NewContact(counterpartyKey.ID.(int64), nil)
					if err := tx.Get(c, txCounterparty.Record); err != nil {
						return err
					}
					if !reflect.DeepEqual(txCounterparty.Data.BalanceJson, counterparty.BalanceJson) {
						return errors.New("contact changed since check: !reflect.DeepEqual(txCounterparty.Balance(), counterparty.Balance())")
					}

					if balanceJson, err := json.Marshal(transfersCounterpartyBalance); err != nil {
						return fmt.Errorf("failed to json.Marshal(transfersCounterpartyBalance): %w", err)
					} else {
						txCounterparty.Data.BalanceJson = string(balanceJson)
						txCounterparty.Data.BalanceCount = len(transfersCounterpartyBalance)
						if err = tx.Set(c, txCounterparty.Record); err != nil {
							return fmt.Errorf("failed to save counterparty with ID=%v: %w", counterpartyID, err)
						}
					}
					return nil
				}, nil)
				if err != nil {
					log.Errorf(c, "Failed to fix counterparty with ID=%v: %v", counterpartyID, err)
				} else {
					log.Infof(c, "Fixed counterparty with ID=%v", counterpartyID)
					//userCounterpartyRecords[i] = txCounterparty.Data
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
