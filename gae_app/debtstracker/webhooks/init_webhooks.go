package webhooks

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func InitWebhooks(router *httprouter.Router) {
	http.HandleFunc("/webhooks/twilio/", TwilioWebhook)
}
