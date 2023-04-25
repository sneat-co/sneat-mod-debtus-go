package api

import (
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"strconv"

	"context"
	"errors"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/facade"
)

func handleSignUpAnonymously(c context.Context, w http.ResponseWriter, r *http.Request) {
	if user, err := dtdal.User.CreateAnonymousUser(c); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	} else {
		if err = SaveUserAgent(c, user.ID, r.UserAgent()); err != nil {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
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
		if err = SaveUserAgent(c, userID, r.UserAgent()); err != nil {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
		ReturnToken(c, w, userID, false, false)
	} else {
		ErrorAsJson(c, w, http.StatusForbidden, errors.New("User is not anonymous."))
	}
}

//func handleLinkOneSignal(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
//	_, err := dtdal.UserOneSignal.SaveUserOneSignal(c, authInfo.UserID, r.PostFormValue("OneSignalUserID"))
//	if err != nil {
//		ErrorAsJson(c, w, http.StatusInternalServerError, err)
//	}
//}
