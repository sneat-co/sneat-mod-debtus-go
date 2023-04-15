package main

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app"
	"github.com/bots-go-framework/bots-host-gae"
	"github.com/strongo/log"
	"google.golang.org/appengine"
)

func main() {
	log.AddLogger(gae.GaeLogger)
	gaeapp.Init(gae.BotHost())
	appengine.Main()
}
