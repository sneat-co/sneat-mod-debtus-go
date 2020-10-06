package main

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/strongo/log"
	"google.golang.org/appengine"
)

func main() {
	log.AddLogger(gaehost.GaeLogger)
	gaeapp.Init(gaehost.GaeBotHost{})
	appengine.Main()
}
