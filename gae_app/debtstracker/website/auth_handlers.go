package website

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/julienschmidt/httprouter"
)

func LoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	q := r.URL.Query()
	userID, err := strconv.ParseInt(q.Get("user"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof(c, "Invalid user parameter")
		return
	}
	secret := q.Get("secret")
	secretItems := strings.Split(secret, ":")
	expirySecStr := secretItems[0]
	log.Infof(c, "expirySeconds: %v; secret: %v", expirySecStr, secret)
	expirySeconds, err := common.DecodeID(expirySecStr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Infof(c, "Failed to decode expiry bytes to seconds")
		return
	}

	expiresAt := time.Unix(expirySeconds, 0)

	expectedSecret := common.SignInt64WithExpiry(c, userID, expiresAt)
	if secret != expectedSecret {
		w.WriteHeader(http.StatusUnauthorized)
		log.Infof(c, "Invalid secret")
		return
	}

	if expiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		log.Infof(c, "expiresAt.Before(time.Now())")
		w.Write([]byte("<html><body style=font-size:xx-large>Your secret has expired. Please generate a new link</body></html>"))
		return
	}

	if _user, err := dal.User.GetUserByID(c, userID); err != nil {
		if db.IsNotFound(err) {
			w.WriteHeader(http.StatusNotFound)
			log.Infof(c, err.Error())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf(c, err.Error())
		}
		return
	} else {
		if _user.EmailAddress != "" {
			log.Infof(c, "_user.EmailAddress: %v", _user.EmailAddress)
		} else {
			gaeUser := user.Current(c)
			if gaeUser == nil {
				log.Infof(c, "appengine.user.Current(): nil")
			} else {
				if gaeUser.Email == "" {
					log.Infof(c, "gaeUser.Email is empty")
				} else {
					log.Infof(c, "gaeUser.Email: %v", gaeUser.Email)
					err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
						u, err := dal.User.GetUserByID(tc, userID)
						if err != nil {
							return errors.Wrap(err, "Failed to load user")
						}
						if u.EmailAddress == "" {
							u.SetEmail(gaeUser.Email, true)
							if err = dal.User.SaveUser(c, u); err != nil {
								err = errors.Wrap(err, "Failed to save user")
							}
						}
						return err
					}, nil)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						log.Errorf(c, err.Error())
					}
				}
			}
		}
	}

	panic("Not implemented")
	//session, _ := common.GetSession(r)
	//session.SetUserID(userID, w)
	//if err = session.Save(r, w); err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	log.Errorf(c, err.Error())
	//	return
	//}

	//w.Write([]byte("<html><body style=font-size:xx-large>User signed</body></html>"))
}
