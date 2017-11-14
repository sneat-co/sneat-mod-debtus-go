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
}

func filterByUserParam(r *http.Request, query *mapper.Query, prop string) (*mapper.Query, error) {
	if query, err := filterByIntParam(r, query, "user", prop); err != nil {
		return query, err
	} else {
		return query, err
	}
}

//func filterByContactParam(r *http.Request, query *mapper.Query, prop string) (*mapper.Query, error) {
//	return filterByIntParam(r, query, "contact", prop)
//}

func filterByIntParam(r *http.Request, query *mapper.Query, param, prop string) (*mapper.Query, error) {
	if pv := r.URL.Query().Get(param); pv != "" {
		if v, err := strconv.ParseInt(pv, 10, 64); err != nil {
			return query, err
		} else if v != 0 {
			query = query.Filter(prop + " =", v)
		}
	}
	return query, nil
}
