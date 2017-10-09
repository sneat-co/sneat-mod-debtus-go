package maintainance

import (
	"github.com/captaincodeman/datastore-mapper"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
)

func RegisterMappers() {
	mapperServer, _ := mapper.NewServer(
		mapper.DefaultPath,
		mapper.DefaultQueue(common.QUEUE_MAPREDUCE),
	)
	http.Handle(mapper.DefaultPath, mapperServer)
	mapper.RegisterJob(&verifyUsers{})
	mapper.RegisterJob(&verifyTransfers{})
	mapper.RegisterJob(&verifyContacts{})
}
