package api

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"golang.org/x/net/context"
	"net/http"
	"github.com/strongo/app/log"
)

func handleSignedWithFacebook(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	log.Debugf(c, "api.handleSignedWithFacebook()")
	fbUserID := r.PostFormValue("fbUserID")
	fbAppID := r.PostFormValue("fbAppID")
	if fbUserID == "" {
		BadRequestMessage(c, w, "fbUserID is missed")
		return
	}
	if fbAppID == "" {
		BadRequestMessage(c, w, "fbAppID is missed")
		return
	}
	user, isNewUser, _, _, _, err := signInFbUser(c, fbAppID, fbUserID, r, authInfo)
	if err != nil {
		authWriteResponseForAuthFailed(c, w, err)
		return
	}
	authWriteResponseForUser(c, w, user, isNewUser)
}
