package apigaedepended

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/log"
	"google.golang.org/appengine/user"
	"github.com/pkg/errors"
)

var handleFunc = http.HandleFunc

func InitApiGaeDepended() {
	handleFunc("/auth/google/signin", dal.HandleWithContext(handleSigninWithGoogle))
	handleFunc("/auth/google/signed", dal.HandleWithContext(handleSignedWithGoogle))
}

const REDIRECT_DESTINATION_PARAM_NAME = "redirect-to"

func handleSigninWithGoogle(c context.Context, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	destinationUrl := query.Get("to")
	if destinationUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing 'to' parameter"))
		return
	}

	callbackUrl := fmt.Sprintf("/auth/google/signed?%v=%v", REDIRECT_DESTINATION_PARAM_NAME, url.QueryEscape(destinationUrl))
	if secret := query.Get("secret"); secret != "" {
		callbackUrl += "&secret=" + secret
	}

	loginUrl, err := user.LoginURL(c, callbackUrl)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	http.Redirect(w, r, loginUrl, http.StatusFound)
}

func handleSignedWithGoogle(c context.Context, w http.ResponseWriter, r *http.Request) {
	var userID int64
	if authInfo, _, err := auth.Authenticate(w, r, false); err != nil {
		if err != auth.ErrNoToken {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	} else {
		userID = authInfo.UserID
	}

	clientInfo := models.NewClientInfoFromRequest(r)

	googleUser := user.Current(c)
	if googleUser == nil {
		err := errors.New("handleSignedWithGoogle(): googleUser == nil")
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	userGoogle, _, err := facade.User.GetOrCreateUserGoogleOnSignIn(c, googleUser, userID, clientInfo)
	if err != nil {
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	if userGoogle.UserGoogleEntity == nil {
		log.Errorf(c, "userGoogle.UserGoogleEntity == nil")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("userGoogle.UserGoogleEntity == nil"))
	}

	log.Debugf(c, "userGoogle.AppUserIntID: %d", userGoogle.AppUserIntID)
	token := auth.IssueToken(userGoogle.AppUserIntID, "web", userGoogle.Email == "alexander.trakhimenok@gmail.com")
	destinationUrl := r.URL.Query().Get(REDIRECT_DESTINATION_PARAM_NAME)

	var delimiter string
	if strings.Contains(destinationUrl, "#") {
		delimiter = "&"
	} else {
		delimiter = "#"
	}
	destinationUrl += delimiter + "signed-in-with=google"
	destinationUrl += "&secret=" + token
	http.Redirect(w, r, destinationUrl, http.StatusFound)
}
