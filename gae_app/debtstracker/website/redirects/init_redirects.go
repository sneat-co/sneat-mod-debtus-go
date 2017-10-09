package redirects

import (
	"github.com/strongo/app"
	"net/http"
	"strconv"
	"fmt"
)

func InitRedirects() {
	http.HandleFunc("/receipt", ReceiptRedirect)

	http.HandleFunc("/transfer",
		RedirectHandlerToEntityPageWithIntID("transfer=%d", "send"))

	http.HandleFunc("/contact",
		RedirectHandlerToEntityPageWithIntID("contact=%d"))

	strongo.AddHttpHandler("/open/new-debt", newDebtRedirect)

	strongo.AddHttpHandler("/choose-currency", chooseCurrencyRedirect)

	strongo.AddHttpHandler("/confirm", confirmEmailRedirect)

}

func RedirectHandlerToEntityPageWithIntID(path string, optionalParams... string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Value of 'id' parameter is not an integer"))
			return
		} else {
			redirectToWebApp(w, r, true, fmt.Sprintf(path, id), nil, optionalParams)
		}
	}
}
