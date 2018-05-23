package appengine

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/log"
)

func init() {
	log.AddLogger(gaehost.GaeLogger)
	gaeapp.Init(gaehost.GaeBotHost{})
}
