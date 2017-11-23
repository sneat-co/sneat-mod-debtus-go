package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"strconv"
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
	registerAsyncJob(&transfersRecreateContacts{})
	registerAsyncJob(&verifyTelegramUserAccounts{})

	http.HandleFunc("/_ah/merge-contacts", mergeContactsHandler)
}

func filterByUserParam(r *http.Request, query *mapper.Query, prop string) (q *mapper.Query, filtered bool, err error) {
	return filterByIntParam(r, query, "user", prop)
}

//func filterByContactParam(r *http.Request, query *mapper.Query, prop string) (*mapper.Query, error) {
//	return filterByIntParam(r, query, "contact", prop)
//}

func filterByIntParam(r *http.Request, query *mapper.Query, param, prop string) (q *mapper.Query, filtered bool, err error) {
	q = query
	if pv := r.URL.Query().Get(param); pv != "" {
		var v int64
		if v, err = strconv.ParseInt(pv, 10, 64); err != nil {
			return
		} else if v != 0 {
			return query.Filter(prop + " =", v), true, nil
		}
	}
	return
}
