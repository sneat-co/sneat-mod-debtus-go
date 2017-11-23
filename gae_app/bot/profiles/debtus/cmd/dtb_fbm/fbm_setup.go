package dtb_fbm

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/fbm"
	"bytes"
	"fmt"
	"github.com/strongo/log"
	"github.com/strongo/bots-api-fbm"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"time"
	"github.com/julienschmidt/httprouter"
)

func SetupFbm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query()

	botID := query.Get("bot")
	c := appengine.NewContext(r)
	bot, ok := fbm.Bots(c).ByCode[botID]
	if !ok {
		w.Write([]byte("Unknown bot: " + botID))
		return
	}

	c, _ = context.WithDeadline(c, time.Now().Add(20*time.Second))
	api := fbm_api.NewGraphApi(urlfetch.Client(c), bot.Token)

	var err error

	var buffer bytes.Buffer

	reportError := func(err error) {
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(buffer.Bytes())
		w.Write([]byte(err.Error()))
	}

	if query.Get("whitelist-domains") == "1" {
		if err = SetWhitelistedDomains(c, r, bot, api); err != nil {
			reportError(err)
			return
		}
		buffer.WriteString("Whitelisted domains\n")
	}

	if query.Get("enable-get-started") == "1" {
		getStartedMessage := fbm_api.GetStartedMessage{}

		getStartedMessage.GetStarted.Payload = "fbm-get-started"

		if err = api.SetGetStarted(c, getStartedMessage); err != nil {
			reportError(err)
			return
		}
		buffer.WriteString("Enabled 'Get started'\n")
	}

	if query.Get("set-persistent-menu") == "1" {
		if err = SetPersistentMenu(c, r, bot, api); err != nil {
			reportError(err)
			return
		}
		buffer.WriteString("Enabled 'Persistent menu'\n")
	}
	log.Debugf(c, buffer.String())
	w.Header().Set("Content-Type", "text/plain")
	w.Write(buffer.Bytes())
	w.Write([]byte(fmt.Sprintf("OK! %v", time.Now())))
}
