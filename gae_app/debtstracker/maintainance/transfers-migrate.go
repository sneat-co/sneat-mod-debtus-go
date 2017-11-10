package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/captaincodeman/datastore-mapper"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
)

type migrateTransfers struct {
	transfersAsyncJob
}

func (m *migrateTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	return m.startTransferWorker(c, counters, key, m.migrateTransfer)
}

func (m *migrateTransfers) migrateTransfer(c context.Context, counters *asyncCounters, transfer models.Transfer) (err error) {
	if !transfer.HasObsoleteProps() {
		return
	}
	if err = datastore.RunInTransaction(c, func(tc context.Context) (err error) {
		if transfer, err = dal.Transfer.GetTransferByID(c, transfer.ID); err != nil {
			return
		}
		if transfer.HasObsoleteProps() {
			if err = dal.Transfer.SaveTransfer(tc, transfer); err != nil {
				return
			}
			log.Infof(c, "Transfer %v fixed", transfer.ID)
		}
		return
	}, nil); err != nil {
		log.Errorf(c, "failed to fix transfer %v: %v", transfer.ID, err)
	}
	return
}
