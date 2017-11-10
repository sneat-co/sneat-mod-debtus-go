package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"strconv"
	"sync"
	"github.com/strongo/app/gaedb"
	"golang.org/x/net/context"
	"github.com/strongo/app/log"
)

func RegisterMappers() {
	mapperServer, _ := mapper.NewServer(
		mapper.DefaultPath,
		mapper.DefaultQueue(common.QUEUE_MAPREDUCE),
	)
	http.Handle(mapper.DefaultPath, mapperServer)

	registerAsyncJob := func(job interface {
		mapper.JobSpec
		mapper.SliceLifecycle
		mapper.JobLifecycle
	}) {
		mapper.RegisterJob(job)
	}
	registerAsyncJob(&verifyUsers{})
	registerAsyncJob(&verifyContacts{})
	registerAsyncJob(&verifyTransfers{})
	registerAsyncJob(&migrateTransfers{})
	registerAsyncJob(&verifyContactTransfers{})
}

func filterByUserParam(r *http.Request, query *mapper.Query, prop string) (*mapper.Query, error) {
	return filterByParam(r, query, "user", prop)
}

func filterByContactParam(r *http.Request, query *mapper.Query, prop string) (*mapper.Query, error) {
	return filterByParam(r, query, "contact", prop)
}

func filterByParam(r *http.Request, query *mapper.Query, param, prop string) (*mapper.Query, error) {
	if pv := r.URL.Query().Get(param); pv != "" {
		if userID, err := strconv.ParseInt(pv, 10, 64); err != nil {
			return query, err
		} else if userID != 0 {
			query = query.Filter(prop + " =", userID)
		}
	}
	return query, nil
}


type asyncMapper struct {
	sync.WaitGroup
	sync.Mutex
}

func (m *asyncMapper) IncrementCounter(counters mapper.Counters, name string, delta int64) {
	m.Lock()
	counters.Increment(name, delta)
	m.Unlock()
}


type processFactory func() func()

func (m *asyncMapper) startProcess(c context.Context, createProcess processFactory) error {
	m.Add(1)
	process := createProcess()
	go func(){
		defer func() {
			m.Done()
			if r := recover(); r != nil {
				log.Errorf(c, "panic: %v", r)
			}
		}()
		process()
	}()
	return nil
}

// JobStarted is called when a mapper job is started
func (asyncMapper) JobStarted(c context.Context, id string) {
	log.Debugf(c, "Job started: %v", id)
}

// JobStarted is called when a mapper job is completed
func (asyncMapper) JobCompleted(c context.Context, id string) {
	logJobCompletion(c, id)
}


func (asyncMapper) SliceStarted(c context.Context, id string, namespace string, shard, slice int) {
	gaedb.LoggingEnabled = false
}

// SliceStarted is called when a mapper job for an individual slice of a
// shard within a namespace is completed
func (m *asyncMapper) SliceCompleted(c context.Context, id string, namespace string, shard, slice int) {
	log.Debugf(c, "Awaiting completion...")
	m.Wait()
	log.Debugf(c, "Processing completed.")
	gaedb.LoggingEnabled = true
}
