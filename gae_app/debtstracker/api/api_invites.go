package api

import (
	"net/http"
	"strconv"

	"context"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"google.golang.org/appengine/user"
)

func CreateInvite(c context.Context, w http.ResponseWriter, r *http.Request) {
	gaeUser := user.Current(c)
	if !gaeUser.Admin {
		w.WriteHeader(http.StatusForbidden)
	}

	createUserData := &dtdal.CreateUserData{}
	clientInfo := models.NewClientInfoFromRequest(r)
	userEmail, _, err := facade.User.GetOrCreateEmailUser(c, gaeUser.Email, true, createUserData, clientInfo)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	_, _ = w.Write([]byte(strconv.FormatInt(userEmail.AppUserIntID, 10)))
}
