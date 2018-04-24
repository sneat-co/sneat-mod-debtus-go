package api

import (
	"net/http"
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"context"
)

// TODO: Obsolete - migrate to handleSignInWithPin
func handleSignInWithCode(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	code := r.PostFormValue("code")
	if code == "" {
		BadRequestMessage(c, w, "Missing required attribute: code")
		return
	}
	if loginCode, err := strconv.Atoi(code); err != nil {
		BadRequestMessage(c, w, "Parameter code is not an integer")
		return
	} else if loginCode == 0 {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Login code should not be 0."))
		return
	} else {
		if userID, err := dal.LoginCode.ClaimLoginCode(c, int32(loginCode)); err != nil {
			switch err {
			case models.ErrLoginCodeExpired:
				w.Write([]byte("expired"))
			case models.ErrLoginCodeAlreadyClaimed:
				w.Write([]byte("claimed"))
			default:
				err = errors.Wrap(err, "Failed to claim code")
				ErrorAsJson(c, w, http.StatusInternalServerError, err)
			}
		} else {
			if authInfo.UserID != 0 && userID != authInfo.UserID {
				log.Warningf(c, "userID:%v != authInfo.AppUserIntID:%v", userID, authInfo.UserID)
			}
			ReturnToken(c, w, userID, false, false)
			return
		}
	}
}

func handleSignInWithPin(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	loginID, err := common.DecodeID(r.PostFormValue("loginID"))
	if err != nil {
		BadRequestError(c, w, errors.Wrap(err, "Parameter 'loginID' is not an integer."))
		return
	}

	if loginCode, err := strconv.ParseInt(r.PostFormValue("loginPin"), 10, 32); err != nil {
		BadRequestMessage(c, w, "Parameter 'loginCode' is not an integer")
		return
	} else if loginCode == 0 {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Parameter 'loginCode' should not be 0."))
		return
	} else {
		if userID, err := facade.AuthFacade.SignInWithPin(c, loginID, int32(loginCode)); err != nil {
			switch err {
			case facade.ErrLoginExpired:
				w.Write([]byte("expired"))
			case facade.ErrLoginAlreadySigned:
				w.Write([]byte("claimed"))
			default:
				err = errors.Wrap(err, "Failed to claim loginCode")
				ErrorAsJson(c, w, http.StatusInternalServerError, err)
			}
		} else {
			if authInfo.UserID != 0 && userID != authInfo.UserID {
				log.Warningf(c, "userID:%v != authInfo.AppUserIntID:%v", userID, authInfo.UserID)
			}
			ReturnToken(c, w, userID, false, false)
		}
	}
}
