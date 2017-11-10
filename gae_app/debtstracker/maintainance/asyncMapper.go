package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"github.com/strongo/app/gaedb"
	"golang.org/x/net/context"
	"runtime/debug"
	"github.com/strongo/app/log"
	"sync"
)

type asyncMapper struct {
	panicked bool
	sync.WaitGroup
}


type Worker func(counters *asyncCounters) error
type WorkerFactory func() Worker

func (m *asyncMapper) startWorker(c context.Context, counters mapper.Counters, createWorker WorkerFactory) error {
	m.panicked = true
	m.Add(1)
	executeWorker := createWorker()
	go func(){
		counters := NewAsynCounters(counters)
		defer func() {
			m.Done()
			if counters.locked {
				counters.Unlock()
			}
			if r := recover(); r != nil {
				log.Errorf(c, "panic: %v\n\tStack trace: %v", r, string(debug.Stack()))
			}
		}()
		if err := executeWorker(counters); err != nil {
			log.Errorf(c, "*contactsAsyncJob() > Worker failed: %v", err)
		}
		m.panicked = false
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

