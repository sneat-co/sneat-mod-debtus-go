package api

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	strongo "github.com/strongo/app"
	"io"
	"net/http"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/log"
)

type AuthHandler func(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo)

type AuthHandlerWithUser func(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser)

func AuthOnly(handler AuthHandler) strongo.ContextHandler {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) {
		log.Debugf(c, "AuthOnly(%v)", handler)
		if authInfo, _, err := auth.Authenticate(w, r, true); err == nil {
			handler(c, w, r, authInfo)
		} else {
			log.Warningf(c, "Failed to authenticate: %v", err.Error())
		}
	}
}

func AuthOnlyWithUser(handler AuthHandlerWithUser) strongo.ContextHandler {
	return AuthOnly(func(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
		var userID int64

		if userID = getUserID(c, w, r, authInfo); userID == 0 {
			log.Warningf(c, "userID is 0")
			return
		}

		user, err := facade.User.GetUserByID(c, nil, userID)

		if hasError(c, w, err, models.AppUserKind, int(userID), http.StatusInternalServerError) {
			return
		}
		handler(c, w, r, authInfo, user)
	})
}

func OptionalAuth(handler AuthHandler) strongo.ContextHandler {
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

func adminOnly(handler AuthHandler) strongo.ContextHandler {
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
		loginID int
		err     error
	)

	loginIdStr := query.Get("id")

	if loginIdStr != "" {
		if loginID, err = common.DecodeIntID(loginIdStr); err != nil {
			BadRequestError(c, w, err)
			return
		}
	}

	returnLoginID := func(loginID int) {
		encoded := common.EncodeIntID(loginID)
		log.Infof(c, "Login ID: %d, Encoded: %v", loginID, encoded)
		if _, err = w.Write([]byte(encoded)); err != nil {
			log.Criticalf(c, "Failed to write login ID to response: %v", err)
		}
	}

	if loginID != 0 {
		if loginPin, err := dtdal.LoginPin.GetLoginPinByID(c, nil, loginID); err != nil {
			if dal.IsNotFound(err) {
				InternalError(c, w, err)
				return
			}
		} else if loginPin.Data.IsActive(channel) {
			returnLoginID(loginID)
			return
		}
	}

	var rBody []byte
	if rBody, err = io.ReadAll(r.Body); err != nil {
		BadRequestError(c, w, fmt.Errorf("failed to read request body: %w", err))
		return
	}
	gaClientID := string(rBody)

	if gaClientID != "" {
		if len(gaClientID) > 100 {
			BadRequestMessage(c, w, fmt.Sprintf("Google Client ID is too long: %d", len(gaClientID)))
			return
		}

		if strings.Count(gaClientID, ".") != 1 {
			BadRequestMessage(c, w, fmt.Sprintf("Google Client ID has wrong format, a '.' char expected"))
			return
		}
	}

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		InternalError(c, w, err)
		return
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
		var loginPin models.LoginPin
		if loginPin, err = dtdal.LoginPin.CreateLoginPin(c, tx, channel, gaClientID, authInfo.UserID); err != nil {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
		loginID = loginPin.ID
		return err
	})
	if err != nil {
		InternalError(c, w, err)
		return
	}
	returnLoginID(loginID)
}
