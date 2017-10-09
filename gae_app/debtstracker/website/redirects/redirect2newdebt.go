package redirects

import "net/http"

func newDebtRedirect(w http.ResponseWriter, r *http.Request) {
	redirectToWebApp(w, r, true, "/main/debts/new-debt", map[string]string{}, []string{})
}
