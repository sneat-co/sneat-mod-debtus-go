package api

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/auth"
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"net/http"
	"strings"
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"io/ioutil"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"github.com/strongo/app/db"
)

type AuthHandler func(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo)

type AuthHandlerWithUser func(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser)

func AuthOnly(handler AuthHandler) dal.ContextHandler {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) {
		log.Debugf(c, "AuthOnly(%v)", handler)
		if authInfo, _, err := auth.Authenticate(w, r, true); err == nil {
			handler(c, w, r, authInfo)
		} else {
			log.Errorf(c, "Failed to authenticate: %v", err.Error())
		}
	}
}

func AuthOnlyWithUser(handler AuthHandlerWithUser) dal.ContextHandler {
	return AuthOnly(func(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
		var userID int64

		if userID = getUserID(c, w, r, authInfo); userID == 0 {
			log.Warningf(c, "userID is 0")
			return
		}

		user, err := dal.User.GetUserByID(c, userID)

		if hasError(c, w, err, models.AppUserKind, userID, http.StatusInternalServerError) {
			return
		}
		handler(c, w, r, authInfo, user)
	})
}

func OptionalAuth(handler AuthHandler) dal.ContextHandler {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) {
		authInfo, _, _ := auth.Authenticate(w, r, false)
		if authInfo.UserID == 0 {
			log.Debugf(c, "OptionalAuth(), anonymous")
		} else {
			log.Debugf(c, "OptionalAuth(), userID=%d", authInfo.UserID)
		}
		handler(c, w, r, authInfo)
	}
}

func adminOnly(handler AuthHandler) dal.ContextHandler {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) {
		log.Debugf(c, "adminOnly(%v)", handler)
		if authInfo, _, err := auth.Authenticate(w, r, true); err == nil {
			if !authInfo.IsAdmin {
				log.Debugf(c, "Not admin!")
				//hashedWriter.WriteHeader(http.StatusForbidden)
				//return
			}
			handler(c, w, r, authInfo)
		} else {
			log.Errorf(c, "Failed to authenticate: %v", err.Error())
		}
	}
}

func IsAdmin(email string) bool {
	return email == "alexander.trakhimenok@gmail.com"
}

func ReturnToken(_ context.Context, w http.ResponseWriter, userID int64, isNewUser, isAdmin bool) {
	token := auth.IssueToken(userID, "api", isAdmin)
	header := w.Header()
	header.Add("Access-Control-Allow-Origin", "*")
	header.Add("Content-Type", "application/json")
	w.Write([]byte("{"))
	if isNewUser {
		w.Write([]byte(`"isNewUser":true,`))
	}
	w.Write([]byte(`"token":"`))
	w.Write([]byte(token))
	w.Write([]byte(`"}`))
}

func handleAuthLoginId(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	query := r.URL.Query()
	channel := query.Get("channel")
	var (
		loginID int64
		err     error
	)

	loginIdStr := query.Get("id")

	if loginIdStr != "" {
		if loginID, err = common.DecodeID(loginIdStr); err != nil {
			BadRequestError(c, w, err)
			return
		}
	}

	returnLoginID := func(loginID int64) {
		encoded := common.EncodeID(loginID)
		log.Infof(c, "Login ID: %d, Encoded: %v", loginID, encoded)
		if _, err = w.Write([]byte(encoded)); err != nil {
			log.Criticalf(c, "Failed to write login ID to response: %v", err)
		}
	}

	if loginID != 0 {
		if loginPin, err := dal.LoginPin.GetLoginPinByID(c, loginID); err != nil {
			if err != db.ErrRecordNotFound {
				InternalError(c, w, err)
				return
			}
		} else if loginPin.IsActive(channel) {
			returnLoginID(loginID)
			return
		}
	}

	var rBody []byte
	if rBody, err = ioutil.ReadAll(r.Body); err != nil {
		BadRequestError(c, w, errors.Wrap(err, "Failed to read request body"))
		return
	}
	gaClientID := string(rBody)

	if gaClientID != "" {
		if len(gaClientID) > 100 {
			BadRequestMessage(c, w,fmt.Sprintf("Google Client ID is too long: %d", len(gaClientID)))
			return
		}

		if strings.Count(gaClientID, ".") != 1 {
			BadRequestMessage(c, w,fmt.Sprintf("Google Client ID has wrong format, a '.' char expected"))
			return
		}
	}

	if loginID, err = dal.LoginPin.CreateLoginPin(c, channel, gaClientID, authInfo.UserID); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	returnLoginID(loginID)
}
