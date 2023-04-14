package api

import (
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"strconv"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"context"
	"errors"
)

func handleSignUpAnonymously(c context.Context, w http.ResponseWriter, r *http.Request) {
	if user, err := dtdal.User.CreateAnonymousUser(c); err != nil {
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

	userEntity, err := facade.User.GetUserByID(c, nil, userID)

	if err != nil {
		if dal.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusBadRequest, err)
		} else {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
		}
		return
	}

	if userEntity.Data.IsAnonymous {
		SaveUserAgent(c, userID, r.UserAgent())
		ReturnToken(c, w, userID, false, false)
	} else {
		ErrorAsJson(c, w, http.StatusForbidden, errors.New("User is not anonymous."))
	}
}

func handleLinkOneSignal(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	_, err := dtdal.UserOneSignal.SaveUserOneSignal(c, authInfo.UserID, r.PostFormValue("OneSignalUserID"))
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	}
}
