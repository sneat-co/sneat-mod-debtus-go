package main

import (
	"github.com/bots-go-framework/bots-host-gae"
	gaeapp "github.com/sneat-co/debtstracker-go/gae_app"
	"github.com/strongo/log"
	"google.golang.org/appengine"
)

func main() {
	log.AddLogger(gae.GaeLogger)
	gaeapp.Init(gae.BotHost())
	appengine.Main()
}
