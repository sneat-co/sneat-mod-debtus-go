package maintainance

import (
	"sync"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
)

type migrateTransfers struct {
	sync.Mutex
	wg     sync.WaitGroup
	entity *models.TransferEntity
}

var _ mapper.JobSpec = (*migrateTransfers)(nil)

func (m *migrateTransfers) Query(r *http.Request) (*mapper.Query, error) {
	query := mapper.NewQuery(models.TransferKind)
	return query, nil
}

func (m *migrateTransfers) Next(c context.Context, counters mapper.Counters, key *datastore.Key) (err error) {
	transferEntity := *m.entity
	if transferEntity.HasObsoleteProps() {
		m.wg.Add(1)
		go m.migrateTransfer(c, counters, models.Transfer{ID: key.IntID(), TransferEntity: &transferEntity})
	}
	return
}

func (m *migrateTransfers) Make() interface{} {
	m.entity = new(models.TransferEntity)
	return m.entity
}

func (m *migrateTransfers) migrateTransfer(c context.Context, counters mapper.Counters, transfer models.Transfer) {
	defer m.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Errorf(c, "panic: %v", r)
		}
	}()
	if err := datastore.RunInTransaction(c, func(tc context.Context) (err error) {
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
}

func (m *migrateTransfers) SliceStarted(c context.Context, id string, namespace string, shard, slice int) {
}

// SliceStarted is called when a mapper job for an individual slice of a
// shard within a namespace is completed
func (m *migrateTransfers) SliceCompleted(c context.Context, id string, namespace string, shard, slice int) {
	log.Debugf(c, "Awaiting completion...")
	m.wg.Wait()
	log.Debugf(c, "Processing completed.")
}
