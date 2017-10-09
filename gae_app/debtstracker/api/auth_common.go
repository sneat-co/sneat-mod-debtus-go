package api

import (
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"golang.org/x/net/context"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/app/log"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/user"
	"github.com/strongo/app/db"
)

func handleDisconnect(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	provider := r.URL.Query().Get("provider")

	if err := dal.DB.RunInTransaction(c, func(c context.Context) error {
		appUser, err := dal.User.GetUserByID(c, authInfo.UserID)
		if err != nil {
			return err
		}

		changed := false

		deleteFbUser := func(userAccount user.Account) error {
			if userFb, err := dal.UserFacebook.GetFbUserByFbID(c, userAccount.App, userAccount.ID); err != nil {
				if err != db.ErrRecordNotFound {
					return err
				}
			} else if userFb.AppUserIntID == appUser.ID {
				if err = dal.UserFacebook.DeleteFbUser(c, userAccount.App, userAccount.ID); err != nil {
					return err
				}
			} else {
				log.Warningf(c, "TODO: Handle case if userFb.AppUserIntID:%d != appUser.ID:%d", userFb.AppUserIntID, appUser.ID)
			}
			return nil
		}

		if !models.IsKnownUserAccountProvider(provider) {
			ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Unknown provider: " + provider))
			return nil
		}
		if !appUser.HasAccount(provider, "") {
			return nil
		}
		var userAccount *user.Account
		switch provider {
		case "google":
			if userAccount, err = appUser.GetGoogleAccount(); err != nil {
				return err
			} else if userAccount != nil {
				if userGoogle, err := dal.UserGoogle.GetUserGoogleByID(c, userAccount.ID); err != nil {
					if err != db.ErrRecordNotFound {
						return err
					}
				} else if userGoogle.AppUserIntID == appUser.ID {
					userGoogle.AppUserIntID = 0
					if err = dal.UserGoogle.DeleteUserGoogle(c, userGoogle.ID); err != nil {
						return err
					}
				} else {
					log.Warningf(c, "TODO: Handle case if userGoogle.AppUserIntID:%d != appUser.ID:%d", userGoogle.AppUserIntID, appUser.ID)
				}
				_ = appUser.RemoveAccount(*userAccount)
				changed = true
			}
		case "fb":
			if userAccount, err = appUser.GetFbAccount(""); err != nil {
				return err
			} else if userAccount != nil {
				if err = deleteFbUser(*userAccount); err != nil {
					return err
				}
				_ = appUser.RemoveAccount(*userAccount)
				changed = true
			}
		case "fbm":
			if userAccount, err = appUser.GetFbAccount(""); err != nil {
				return err
			} else if userAccount != nil {
				if err = deleteFbUser(*userAccount); err != nil {
					return err
				}
				_ = appUser.RemoveAccount(*userAccount)
				changed = true
			}
		default:
		}

		if changed {
			if err = dal.User.SaveUser(c, appUser); err != nil {
				return err
			}
		}
		return nil
	}, dal.CrossGroupTransaction); err != nil {
		ErrorAsJson(c,  w, http.StatusInternalServerError, err)
	}
}
