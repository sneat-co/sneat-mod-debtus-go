package api

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/user"
	"net/http"
	"strconv"
	"bitbucket.com/debtstracker/gae_app/debtstracker/facade"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
)

func ApiCreateInvite(c context.Context, w http.ResponseWriter, r *http.Request) {
	gaeUser := user.Current(c)
	if !gaeUser.Admin {
		w.WriteHeader(http.StatusForbidden)
	}

	createUserData := &dal.CreateUserData{}
	clientInfo := models.NewClientInfoFromRequest(r)
	userEmail, _, err := facade.User.GetOrCreateEmailUser(c, gaeUser.Email, true, createUserData, clientInfo)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	w.Write([]byte(strconv.FormatInt(userEmail.AppUserIntID, 10)))
	return
}
