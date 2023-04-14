package api

import (
	"net/http"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

func handleGetUserCurrencies(c context.Context, w http.ResponseWriter, _ *http.Request, _ auth.AuthInfo, user models.AppUser) {
	jsonToResponse(c, w, user.Data.LastCurrencies)
}
