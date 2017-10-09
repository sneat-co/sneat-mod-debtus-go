package api

import (
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"golang.org/x/net/context"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
)

func handleGetUserCurrencies(c context.Context, w http.ResponseWriter, _ *http.Request, _ auth.AuthInfo, user models.AppUser) {
	jsonToResponse(c, w, user.LastCurrencies)
}
