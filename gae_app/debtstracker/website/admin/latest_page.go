package admin

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal/gaedal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bufio"
	"fmt"
	"github.com/strongo/app/log"
	"golang.org/x/net/html"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	gaeUser "google.golang.org/appengine/user"
	"net/http"
	"strconv"
)

func LatestPage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if !gaeUser.IsAdmin(c) {
		url, _ := gaeUser.LoginURL(c, r.RequestURI)
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Header().Add("location", url)
		return
	}

	var users []models.AppUserEntity
	userKeys, err := datastore.NewQuery(models.AppUserKind).Order("-DtCreated").Limit(50).GetAll(c, &users)
	if err != nil {
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	b := bufio.NewWriter(w)
	b.WriteString("<html><head><style>body{Font-Family:Verdana;font-size:x-small} td{padding: 2px 5px; background-color: #eee} th{padding: 2px 5px; text-align: left; background-color: #ddd} .num{text-align: right} div{float: left}</style></head>")
	b.WriteString("<body><h1>Latest</h1><hr>")
	b.WriteString("<div><h2>Users</h2><table cellspacing=1><thead><tr><th>#</th><th>Full ContactName</th><th>Contacts</th><th>Debts</th><th>Balance</th><th>Invited by</th></tr></thead><tbody>")
	for i, user := range users {
		b.WriteString("<tr>")
		b.WriteString("<td class=num>")
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString("</td><td>")
		b.WriteString(fmt.Sprintf("<a href='user?id=%v'>%v</a>", userKeys[i].IntID(), html.EscapeString(user.FullName())))
		b.WriteString("</td><td class=num>")
		b.WriteString(strconv.Itoa(user.TotalContactsCount()))
		b.WriteString("</td><td class=num>")
		b.WriteString(strconv.Itoa(user.CountOfTransfers))
		b.WriteString("</td><td>")
		b.WriteString(user.BalanceJson)
		b.WriteString("</td><td>")
		if user.InvitedByUserID != 0 {
			if invitedByUser, err := dal.User.GetUserByID(c, user.InvitedByUserID); err != nil {
				b.WriteString(err.Error())
			} else {
				b.WriteString(fmt.Sprintf("<a href='user?id=%v>%v</a>')", user.InvitedByUserID, invitedByUser.FullName()))
			}
		}
		b.WriteString("</td></tr>")
	}
	b.WriteString("</tbody></table></div>")
	b.WriteString("</body></html>")
	b.Flush()
}

func FixTransfersHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	loadedCount, fixedCount, failedCount, err := gaedal.FixTransfers(c)
	stats := fmt.Sprintf("\nLoaded: %v, Fixed: %v, Failed: %v", loadedCount, fixedCount, failedCount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Write([]byte(stats))
}
