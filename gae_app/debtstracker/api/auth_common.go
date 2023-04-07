package api

import (
	"net/http"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/app/user"
	"github.com/strongo/log"
)

func handleDisconnect(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	provider := r.URL.Query().Get("provider")

	if err := dtdal.DB.RunInTransaction(c, func(c context.Context) error {
		appUser, err := facade.User.GetUserByID(c, authInfo.UserID)
		if err != nil {
			return err
		}

		changed := false

		deleteFbUser := func(userAccount user.Account) error {
			if userFb, err := dtdal.UserFacebook.GetFbUserByFbID(c, userAccount.App, userAccount.ID); err != nil {
				if err != db.ErrRecordNotFound {
					return err
				}
			} else if userFb.AppUserIntID == appUser.ID {
				if err = dtdal.UserFacebook.DeleteFbUser(c, userAccount.App, userAccount.ID); err != nil {
					return err
				}
			} else {
				log.Warningf(c, "TODO: Handle case if userFb.AppUserIntID:%d != appUser.ID:%d", userFb.AppUserIntID, appUser.ID)
			}
			return nil
		}

		if !models.IsKnownUserAccountProvider(provider) {
			ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Unknown provider: "+provider))
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
				if userGoogle, err := dtdal.UserGoogle.GetUserGoogleByID(c, userAccount.ID); err != nil {
					if err != db.ErrRecordNotFound {
						return err
					}
				} else if userGoogle.AppUserIntID == appUser.ID {
					userGoogle.AppUserIntID = 0
					if err = dtdal.UserGoogle.DeleteUserGoogle(c, userGoogle.ID); err != nil {
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
			if err = facade.User.SaveUser(c, appUser); err != nil {
				return err
			}
		}
		return nil
	}, dtdal.CrossGroupTransaction); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
	}
}
