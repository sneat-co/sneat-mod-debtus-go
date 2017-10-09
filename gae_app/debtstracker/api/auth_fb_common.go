package api

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/fbm"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"github.com/strongo/app/log"
	fb "github.com/strongo/facebook"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"time"
)

var ErrUnauthorized = errors.New("Unauthorized")
var ErrBadRequest = errors.New("Bad request")

func signInFbUser(c context.Context, fbAppID, fbUserID string, r *http.Request, authInfo auth.AuthInfo) (
	user models.AppUser, isNewUser bool, userFacebook models.UserFacebook, fbApp *fb.App, fbSession *fb.Session, err error,
) {
	log.Debugf(c, "api.signInFbUser()")

	if fbAppID == "" {
		panic("fbAppID is empty string")
		return
	}
	if fbUserID == "" {
		panic("fbUserID is empty string")
		return
	}

	signedRequest := r.PostFormValue("signed_request")
	accessToken := r.PostFormValue("access_token")

	var isFbm bool

	// Create FB API Session
	{
		if signedRequest != "" && accessToken != "" {
			err = errors.WithMessage(ErrBadRequest, "Parameters access_token & signed_request should not be passed together")
			return
		} else if accessToken != "" {
			_, fbSession, err = fbm.FbAppAndSessionFromAccessToken(c, r, accessToken)
		} else if signedRequest != "" {
			var (
				signedData fb.Result
			)
			if fbApp, _, err = fbm.GetFbAppAndHost(r); err != nil {
				return
			}
			if signedData, err = fbApp.ParseSignedRequest(signedRequest); err != nil {
				return
			}
			psID := signedData.Get("psid").(string)
			if psID != "" {
				if fbUserID == "" {
					fbUserID = psID
				} else if fbUserID != psID {
					err = errors.WithMessage(ErrBadRequest, "fbUserID != psID")
					return
				}
				var (
					pageID string
					ok     bool
				)
				if pageID, ok = signedData.Get("page_id").(string); !ok {
					pageID = strconv.FormatFloat(signedData.Get("page_id").(float64), 'f', 0, 64)
				}

				log.Debugf(c, "pageID: %v, signedData: %v", pageID, signedData)
				if fbmBot, ok := fbm.Bots(c).ByID[pageID]; !ok {
					err = errors.New("Bot settings not found by page ID=" + pageID)
				} else {
					isFbm = true
					_, fbSession, err = fbm.FbAppAndSessionFromAccessToken(c, r, fbmBot.Token)
				}
			} else {
				err = fmt.Errorf("Not implemented for signed request: %v", signedData)
				return
			}
		} else {
			err = errors.WithMessage(ErrBadRequest, "Either access_token or signed_request should be passed")
			return
		}
		if err != nil {
			err = errors.WithMessage(ErrUnauthorized, err.Error())
			return
		}
	}

	if userFacebook, err = dal.UserFacebook.GetFbUserByFbID(c, fbAppID, fbUserID); err != nil && !db.IsNotFound(err) {
		err = errors.WithMessage(err, "Failed to get UserFacebook record by ID")
		return
	} else if !db.IsNotFound(err) && fbUserID != "" && fbUserID != userFacebook.FbUserOrPageScopeID {
		err = errors.WithMessage(ErrUnauthorized, fmt.Sprintf("fbUserID:%v != userFacebook.ID:%v", fbUserID, userFacebook.FbUserOrPageScopeID))
		return
	}

	if accessToken != "" || userFacebook.UserFacebookEntity == nil || userFacebook.DtUpdated.Before(time.Now().Add(-1*time.Hour)) {
		if user, userFacebook, isNewUser, err = createOrUpdateFbUserDbRecord(c, isFbm, fbAppID, fbUserID, fbSession, authInfo, models.NewClientInfoFromRequest(r)); err != nil {
			return
		}
	} else {
		log.Debugf(c, "Not updating FB user db record as last updated less then an hour ago")
	}

	if err != nil {
		return
	} else if user.ID == 0 {
		panic("userID == 0")
	} else if user.AppUserEntity == nil {
		panic("user.AppUserEntity == nil")
	}

	return
}

func getFbUserInfo(c context.Context, fbSession *fb.Session, isFbm bool, fbUserID string,
) (
	emailConfirmed bool, email, firstName, lastName string, err error,
) {
	var (
		endPoint string
		fields   string
	)
	if isFbm {
		endPoint = "/" + fbUserID
		fields = "first_name,last_name,profile_pic,locale,timezone,gender,is_payment_enabled,last_ad_referral"
	} else {
		endPoint = "/me"
		fields = "email,first_name,last_name" //TODO: Try to add fields matching FBM case above. profile_pic is not OK :(
	}
	fbResp, err := fbSession.Get(endPoint, fb.Params{
		"fields": fields,
	})

	if err != nil {
		err = errors.WithMessage(err, "Failed to call Facebook API")
		return
	}

	log.Debugf(c, "Facebook API response: %v", fbResp)

	var ok bool
	if email, ok = fbResp["email"].(string); ok && email != "" {
		emailConfirmed = true
	} else {
		email = fmt.Sprintf("%v@fb", fbUserID)
	}

	firstName, _ = fbResp["first_name"].(string)
	lastName, _ = fbResp["last_name"].(string)
	return
}

func createOrUpdateFbUserDbRecord(c context.Context, isFbm bool, fbAppID, fbUserID string, fbSession *fb.Session, authInfo auth.AuthInfo, clientInfo models.ClientInfo) (user models.AppUser, userFacebook models.UserFacebook, isNewUser bool, err error) {
	var (
		emailConfirmed             bool
		email, firstName, lastName string
	)
	emailConfirmed, email, firstName, lastName, err = getFbUserInfo(c, fbSession, isFbm, fbUserID)

	userFacebook, user, err = facade.User.GetOrCreateUserFacebookOnSignIn(c, authInfo.UserID, fbAppID, fbUserID, firstName, lastName, email, emailConfirmed, clientInfo)
	if err != nil {
		return
	}
	return
}

func authWriteResponseForAuthFailed(c context.Context, w http.ResponseWriter, err error) {
	if errors.Cause(err) == ErrUnauthorized {
		w.WriteHeader(http.StatusUnauthorized)
		log.Debugf(c, err.Error())
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(c, "Auth failed: %v", err.Error())
	}
	w.Write([]byte(err.Error()))
}

func authWriteResponseForUser(c context.Context, w http.ResponseWriter, user models.AppUser, isNewUser bool) {
	ReturnToken(c, w, user.ID, isNewUser, user.EmailConfirmed && IsAdmin(user.EmailAddress))
}
