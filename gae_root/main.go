package appengine

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/log"
)

func init() {
	log.AddLogger(gaehost.GaeLogger)
	gae_app.Init(gaehost.GaeBotHost{})
}
