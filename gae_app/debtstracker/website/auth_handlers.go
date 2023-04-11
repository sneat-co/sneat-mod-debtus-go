package website

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/v2/user"
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

	if _user, err := facade.User.GetUserByID(c, tx, userID); err != nil {
		if dal.IsNotFound(err) {
			w.WriteHeader(http.StatusNotFound)
			log.Infof(c, err.Error())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf(c, err.Error())
		}
		return
	} else {
		if _user.Data.EmailAddress != "" {
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
					err = dtdal.DB.RunInTransaction(c, func(tc context.Context) error {
						u, err := facade.User.GetUserByID(tc, tx, userID)
						if err != nil {
							return err
						}
						if u.Data.EmailAddress == "" {
							u.Data.SetEmail(gaeUser.Email, true)
							if err = facade.User.SaveUser(c, tx, u); err != nil {
								return fmt.Errorf("failed to save user: %w", err)
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
