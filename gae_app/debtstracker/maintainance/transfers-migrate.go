package maintainance

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/strongo/log"
	"google.golang.org/appengine/datastore"
)

type migrateTransfers struct {
	transfersAsyncJob
}

func (m *migrateTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	return m.startTransferWorker(c, counters, key, m.migrateTransfer)
}

func (m *migrateTransfers) migrateTransfer(c context.Context, counters *asyncCounters, transfer models.Transfer) (err error) {
	if transfer.CreatorUserID == 0 {
		log.Errorf(c, "Transfer(ID=%v) is missing CreatorUserID")
		return
	}
	if !transfer.HasObsoleteProps() {
		// log.Debugf(c, "transfer.ID=%v has no obsolete props", transfer.ID)
		return
	}
	if err = datastore.RunInTransaction(c, func(tc context.Context) (err error) {
		if transfer, err = facade.GetTransferByID(c, transfer.ID); err != nil {
			return
		}
		if transfer.HasObsoleteProps() {
			if err = facade.Transfers.SaveTransfer(tc, transfer); err != nil {
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
