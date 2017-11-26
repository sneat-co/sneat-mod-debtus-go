package pages

import (
	"fmt"
	"net/http"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/log"
	"golang.org/x/net/html"
	"google.golang.org/appengine"
)

func CounterpartyPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)
	log.Infof(c, "CounterpartyPage: %v", r.Method)
	encodedCounterpartyID := r.URL.Query().Get("id")
	counterpartyID, err := common.DecodeID(encodedCounterpartyID)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	counterparty, err := dal.Contact.GetContactByID(c, counterpartyID)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(fmt.Sprintf(`<html>
	<head><title>Contact: %v</title>
	<meta name="description" content="Transfered amount: %v">
	<link rel="canonical" href="./counterparty?id=%v" />
	<style>
	body{padding: 50px; font-family: Verdana; font-size: small;}
	th{padding-right:10px;text-align:left;}
	</style>
	</head>
	<body>
	<header><a href="/">DebtsTracker.io</a></header>
	<hr>
	<h1>Contact: %v</h1>

	<footer style="margin-top:50px; border-top: 1px solid lightgrey; padding-top:10px">
	<small style="color:grey">2016 &copy; Powered by <a href="https://golang.org/" target="_blank">Go lang</a> & <a href="https://cloud.google.com/appengine/" target="_blank">AppEngine</a></small>
	</footer>
	%v
	</body></html>`, html.EscapeString(counterparty.FullName()), encodedCounterpartyID, html.EscapeString(counterparty.FullName()), html.EscapeString(counterparty.FullName()), GA_CODE)))
}
