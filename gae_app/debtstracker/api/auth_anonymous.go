package api

import (
	"net/http"
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"context"
)

func handleSignUpAnonymously(c context.Context, w http.ResponseWriter, r *http.Request) {
	if user, err := dal.User.CreateAnonymousUser(c); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	} else {
		SaveUserAgent(c, user.ID, r.UserAgent())
		ReturnToken(c, w, user.ID, true, false)
	}
}

func handleSignInAnonymous(c context.Context, w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PostFormValue("user"), 10, 64)
	if err != nil {
		ErrorAsJson(c, w, http.StatusBadRequest, err)
		return
	}

	userEntity, err := dal.User.GetUserByID(c, userID)

	if err != nil {
		if db.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusBadRequest, err)
		} else {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
		}
		return
	}

	if userEntity.IsAnonymous {
		SaveUserAgent(c, userID, r.UserAgent())
		ReturnToken(c, w, userID, false, false)
	} else {
		ErrorAsJson(c, w, http.StatusForbidden, errors.New("User is not anonymous."))
	}
}

func handleLinkOneSignal(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	_, err := dal.UserOneSignal.SaveUserOneSignal(c, authInfo.UserID, r.PostFormValue("OneSignalUserID"))
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	}
}
