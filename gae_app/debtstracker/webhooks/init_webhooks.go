package webhooks

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func InitWebhooks(router *httprouter.Router) {
	http.HandleFunc("/webhooks/twilio/", TwilioWebhook)
}
