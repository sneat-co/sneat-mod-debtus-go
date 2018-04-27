package maintainance

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"

	"context"
	"github.com/captaincodeman/datastore-mapper"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type asyncMapper struct {
	WaitGroup *sync.WaitGroup
}

type Worker func(counters *asyncCounters) error
type WorkerFactory func() Worker

func (m *asyncMapper) startWorker(c context.Context, counters mapper.Counters, createWorker WorkerFactory) (err error) {
	//gaedb.LoggingEnabled = false
	//log.Debugf(c, "*asyncMapper.startWorker()")
	executeWorker := createWorker()
	//log.Debugf(c, "Will add 1 to WaitGroup")
	m.WaitGroup.Add(1)
	//log.Debugf(c, "Added 1 to WaitGroup")
	go func() {
		//log.Debugf(c, "*asyncMapper.startWorker() => goroutine started")
		defer m.WaitGroup.Done()
		counters := NewAsynCounters(counters)
		defer func() {
			if r := recover(); r != nil {
				//gaedb.LoggingEnabled = true
				log.Errorf(c, "panic: %v\n\tStack trace: %v", r, string(debug.Stack()))
				//gaedb.LoggingEnabled = false
			}
			if counters != nil && counters.locked {
				counters.Unlock()
			}
		}()
		if err = executeWorker(counters); err != nil {
			//gaedb.LoggingEnabled = true
			log.Errorf(c, "*contactsAsyncJob() > Worker failed: %v", err)
			//gaedb.LoggingEnabled = false
		}
		//log.Debugf(c, "worker completed")
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

func (m *asyncMapper) SliceStarted(c context.Context, id string, namespace string, shard, slice int) {
	if m.WaitGroup == nil {
		m.WaitGroup = new(sync.WaitGroup)
	}
	//gaedb.LoggingEnabled = false
}

// SliceStarted is called when a mapper job for an individual slice of a
// shard within a namespace is completed
func (m *asyncMapper) SliceCompleted(c context.Context, id string, namespace string, shard, slice int) {
	log.Debugf(c, "Awaiting completion...")
	if m.WaitGroup != nil {
		m.WaitGroup.Wait()
	}
	log.Debugf(c, "Processing completed.")
	//gaedb.LoggingEnabled = true
}

func filterByIntID(r *http.Request, kind, paramName string) (query *mapper.Query, filtered bool, err error) {
	query = mapper.NewQuery(kind)
	paramVal := r.URL.Query().Get(paramName)
	if paramVal == "" {
		return
	}
	var id int64
	if id, err = strconv.ParseInt(paramVal, 10, 64); err != nil {
		err = errors.WithMessage(err, "failed to filter by ID")
		return
	}
	c := appengine.NewContext(r)
	query = query.Filter("__key__ =", datastore.NewKey(c, kind, "", id, nil))
	log.Debugf(c, "Filtered by %v(IntID=%v)", kind, id)
	filtered = true
	return
}

func filterByStrID(r *http.Request, kind, paramName string) (query *mapper.Query, filtered bool, err error) {
	query = mapper.NewQuery(kind)
	paramVal := r.URL.Query().Get(paramName)
	if paramVal == "" {
		return
	}
	c := appengine.NewContext(r)
	query = query.Filter("__key__ =", datastore.NewKey(c, kind, paramVal, 0, nil))
	log.Debugf(c, "Filtered by %v(StrID=%v)", kind, paramVal)
	filtered = true
	return
}
