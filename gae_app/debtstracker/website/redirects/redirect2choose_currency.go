package redirects

import "net/http"

func chooseCurrencyRedirect(w http.ResponseWriter, r *http.Request) {
	redirectToWebApp(w, r, true, "/choose-currency/", map[string]string{}, []string{})
}
