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
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/user"
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

	if googleUser, _, err := facade.User.GetOrCreateUserGoogleOnSignIn(c, user.Current(c), userID, clientInfo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		log.Debugf(c, "googleUser.AppUserIntID: %d", googleUser.AppUserIntID)
		isAdmin := googleUser.Email == "alexander.trakhimenok@gmail.com"
		token := auth.IssueToken(googleUser.AppUserIntID, "web", isAdmin)
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
}
