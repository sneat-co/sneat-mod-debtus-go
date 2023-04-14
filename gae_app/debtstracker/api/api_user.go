package api

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/api/dto"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal/gaedal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/log"
)

func getApiUser(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) (user models.AppUser, err error) {
	if user.ID = getUserID(c, w, r, authInfo); user.ID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user, err = facade.User.GetUserByID(c, nil, user.ID); hasError(c, w, err, models.AppUserKind, int(user.ID), 0) {
		return
	} else if user.Data == nil {
		w.Write([]byte(fmt.Sprintf("User not found by ID=%v", user.ID)))
		http.NotFound(w, r) // TODO: Check response output
		return
	}
	return
}

func handleUserInfo(c context.Context, w http.ResponseWriter, r *http.Request) {
	if userID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(([]byte)(err.Error()))
	} else {
		if err := SaveUserAgent(c, userID, r.UserAgent()); err != nil {
			log.Errorf(c, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(([]byte)(err.Error()))
		}
	}
}

func SaveUserAgent(c context.Context, userID int64, userAgent string) error {
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return nil
	}
	_, err := dtdal.UserBrowser.SaveUserBrowser(c, userID, userAgent)
	return err
}

func handleSaveVisitorData(c context.Context, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gaClientId := r.FormValue("gaClientId")
	if gaClientId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userAgent := r.UserAgent()
	ipAddress := strings.SplitN(r.RemoteAddr, ":", 1)[0]

	if _, err := dtdal.UserGaClient.SaveGaClient(c, gaClientId, userAgent, ipAddress); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	}
}

func handleMe(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser) {
	meDto := dto.UserMeDto{
		UserID:   strconv.FormatInt(authInfo.UserID, 10),
		FullName: user.Data.FullName(),
	}
	if ua, err := user.Data.GetGoogleAccount(); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	} else if ua != nil {
		meDto.GoogleUserID = ua.ID
	}

	if fbAccounts, err := user.Data.GetFbAccounts(); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	} else {
		for _, ua := range fbAccounts {
			meDto.FbUserID = ua.ID
			break // TODO: change to return map of IDs.
		}
	}

	if meDto.FullName == models.NoName {
		meDto.FullName = ""
	}

	jsonToResponse(c, w, meDto)
}

func setUserName(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	if len(body) == 0 {
		ErrorAsJson(c, w, http.StatusBadRequest, fmt.Errorf("%w: User name is required", ErrBadRequest))
		return
	}

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		user, err := facade.User.GetUserByID(c, tx, authInfo.UserID)
		if err != nil {
			return err
		}
		user.Data.Username = string(body)
		if err = facade.User.SaveUser(c, tx, user); err != nil {
			return err
		}
		if err = gaedal.DelayUpdateTransfersWithCreatorName(c, user.ID); err != nil {
			return err
		}
		return err
	})

	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}
