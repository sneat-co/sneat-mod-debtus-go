package main

import (
	"github.com/bots-go-framework/bots-host-gae"
	gaeapp "github.com/sneat-co/sneat-mod-debtus-go/gae_app"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2"
)

func main() {
	log.AddLogger(gae.GaeLogger)
	gaeapp.Init(gae.BotHost())
	appengine.Main()
}
