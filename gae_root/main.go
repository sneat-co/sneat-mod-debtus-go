package appengine

import (
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/hosts/appengine"
	"bitbucket.com/asterus/debtstracker-server/gae_app"
)

func init() {
	log.AddLogger(gae_host.GaeLogger)
	gae_app.Init(gae_host.GaeBotHost{})
}
