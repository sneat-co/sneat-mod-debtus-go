package maintainance

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
)

func RegisterMappers() {
	mapperServer, _ := mapper.NewServer(
		mapper.DefaultPath,
		mapper.DefaultQueue(common.QUEUE_MAPREDUCE),
	)
	http.Handle(mapper.DefaultPath, mapperServer)
	mapper.RegisterJob(&verifyUsers{})
	mapper.RegisterJob(&verifyContacts{})
	mapper.RegisterJob(&verifyTransfers{})
	mapper.RegisterJob(&migrateTransfers{})
	mapper.RegisterJob(&verifyContactTransfers{})
}
