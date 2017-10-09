package appengine

import (
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/hosts/appengine"
	"bitbucket.com/debtstracker/gae_app"
)

func init() {
	log.AddLogger(gae_host.GaeLogger)
	gae_app.Init(gae_host.GaeBotHost{})
}
