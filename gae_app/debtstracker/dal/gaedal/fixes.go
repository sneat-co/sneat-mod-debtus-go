package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	//"github.com/pkg/errors"
	"sync"

	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type TransferFixter struct {
	changed     bool
	Fixes       []string
	transferKey *datastore.Key
	transfer    *models.TransferEntity
}

func NewTransferFixter(transferKey *datastore.Key, transfer *models.TransferEntity) TransferFixter {
	return TransferFixter{transferKey: transferKey, transfer: transfer, Fixes: make([]string, 0)}
}

func (f *TransferFixter) needFixCounterpartyCounterpartyName() bool {
	return f.transfer.Creator().ContactName == ""
}

//func (f *TransferFixter) fixCounterpartyCounterpartyName(c context.Context) error {
//	if f.needFixCounterpartyCounterpartyName() {
//		log.Debugf(c, "%v: needFixCounterpartyCounterpartyName=true", f.transferKey.IntegerID())
//		if f.transfer.Creator().CounterpartyID != 0 {
//			var counterpartyCounterparty models.ContactEntity
//			err := gaedb.Get(c, NewCounterpartyKey(c, f.transfer.Creator().CounterpartyID), &counterpartyCounterparty)
//			if err != nil {
//				return err
//			}
//			f.transfer.Creator().ContactName = counterpartyCounterparty.FullName()
//			log.Debugf(c, "%v: got name from counterpartyCounterparty", f.transferKey.IntegerID())
//			if f.transfer.Creator().ContactName == "" {
//				log.Warningf(c, "Counterparty %v has no full name", f.transfer.Creator().CounterpartyID)
//			}
//		}
//		if f.transfer.Creator().ContactName == "" { // Not fixed from counterparty
//			user, err := dal.User.GetUserByID(c, f.transfer.CreatorUserID)
//			if err != nil {
//				return err
//			}
//			f.transfer.Creator().ContactName = user.FullName()
//			log.Debugf(c, "%v: got name from user", f.transferKey.IntegerID())
//			if f.transfer.Creator().ContactName == "" {
//				log.Warningf(c, "User %v has no full name", f.transfer.CreatorUserID)
//			}
//		}
//		if f.transfer.Creator().ContactName == "" {
//			return errors.New("f.transfer.Creator().ContactName is not fixed")
//		}
//		f.changed = true
//		f.Fixes = append(f.Fixes, "CounterpartyCounterpartyName")
//		//} else {
//		//	log.Debugf(c, "%v: %v", f.transferKey.IntegerID(), f.transfer.Creator().ContactName)
//	}
//	return nil
//}

func (f *TransferFixter) needFixes(c context.Context) bool {
	return f.needFixCounterpartyCounterpartyName()
	//log.Debugf(c, "%v: needFixes=%v", f.transferKey.IntegerID(), result)
	//return result
}

func (f *TransferFixter) FixAllIfNeeded(c context.Context) (err error) {
	if f.needFixes(c) {
		err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
			transfer, err := dal.Transfer.GetTransferByID(tc, f.transferKey.IntID())
			if err != nil {
				return err
			}
			f.transfer = transfer.TransferEntity
			//if err = f.fixCounterpartyCounterpartyName(c); err != nil {
			//	return err
			//}
			if f.changed {
				//log.Debugf(c, "%v: changed", f.transferKey.IntegerID())
				_, err = gaedb.Put(tc, f.transferKey, f.transfer)
				return err
				//} else {
				//	log.Debugf(c, "%v: not changed", f.transferKey.IntegerID())
			}
			return nil
		}, nil)
	}
	return
}

func FixTransfers(c context.Context) (loadedCount int, fixedCount int, failedCount int, err error) {
	query := datastore.NewQuery(models.TransferKind) //.Limit(50)
	iterator := query.Run(c)
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	for {
		var (
			transfer    models.TransferEntity
			transferKey *datastore.Key
		)
		if transferKey, err = iterator.Next(&transfer); err != nil {
			if err == datastore.Done {
				err = nil
				return
			}
			log.Errorf(c, "Failed to get next transfer: %v", err.Error())
			return
		}
		loadedCount += 1
		wg.Add(1)
		go func(transferKey *datastore.Key, transfer models.TransferEntity) {
			defer wg.Done()
			fixter := NewTransferFixter(transferKey, &transfer)
			err2 := fixter.FixAllIfNeeded(c)
			if err2 != nil {
				log.Errorf(c, "Faield to fix transfer=%v: %v", transferKey.IntID(), err2.Error())
				mutex.Lock()
				failedCount += 1
				err = err2
				mutex.Unlock()
			} else {
				if len(fixter.Fixes) > 0 {
					mutex.Lock()
					fixedCount += 1
					mutex.Unlock()
					log.Infof(c, "Fixed transfer %v: %v", transferKey.IntID(), fixter.Fixes)
					//} else {
					//	log.Debugf(c, "Transfer %v is OK: CounterpartyCounterpartyName: %v", transferKey.IntegerID(), fixter.transfer.Creator().ContactName)
				}
			}
		}(transferKey, transfer)
		if err != nil {
			break
		}
	}
	wg.Wait()
	return
}
