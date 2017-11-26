package appengine

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/log"
)

func init() {
	log.AddLogger(gae_host.GaeLogger)
	gae_app.Init(gae_host.GaeBotHost{})
}
