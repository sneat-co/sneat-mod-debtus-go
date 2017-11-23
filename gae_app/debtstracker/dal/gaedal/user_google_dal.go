package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func NewUserGoogleKey(c context.Context, id string) *datastore.Key {
	return gaedb.NewKey(c, models.UserGoogleKind, id, 0, nil)
}

type UserGoogleDalGae struct {
}

func NewUserGoogleDalGae() UserGoogleDalGae {
	return UserGoogleDalGae{}
}

func (_ UserGoogleDalGae) GetUserGoogleByID(c context.Context, googleUserID string) (userGoogle models.UserGoogle, err error) {
	userGoogle.ID = googleUserID
	userGoogle.UserGoogleEntity = new(models.UserGoogleEntity)
	if err = gaedb.Get(c, NewUserGoogleKey(c, googleUserID), userGoogle.UserGoogleEntity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.ErrRecordNotFound
		}
		return
	}
	return
}

func (_ UserGoogleDalGae) DeleteUserGoogle(c context.Context, googleUserID string) (err error) {
	if err = gaedb.Delete(c, NewUserGoogleKey(c, googleUserID)); err != nil {
		return
	}
	return
}

func (_ UserGoogleDalGae) SaveUserGoogle(c context.Context, userGoogle models.UserGoogle) (err error) {
	if _, err = gaedb.Put(c, NewUserGoogleKey(c, userGoogle.ID), userGoogle.UserGoogleEntity); err != nil {
		return
	}
	return
}

// TODO: Obsolete!
//func (_ UserGoogleDalGae) CreateUserGoogle(c context.Context, user user.User, appUserID int64, onSignIn bool, userAgent, remoteAddr string) (entity *models.UserGoogleEntity, isNewGoogleUser, isNewAppUser bool, err error) {
//	err = dal.DB.RunInTransaction(c, func(tc context.Context) (err error) {
//		key := NewUserGoogleKey(tc, user.ID)
//		entity = new(models.UserGoogleEntity)
//
//		if err = gaedb.Get(tc, key, entity); err == nil {
//			if onSignIn {
//				entity.LastSignIn = time.Now()
//
//				if appUserID != 0 && entity.AppUserIntID != appUserID { // Reconnect Google account to different user
//
//					if entity.AppUserIntID == 0 {
//						if appUser, err := dal.User.GetUserByID(c, appUserID); err != nil {
//							return err
//						} else /* if appUser.GoogleUniqueUserID == "" */ {
//							appUser.GoogleUniqueUserID = user.ID
//							if err = dal.User.SaveUser(c, appUser); err != nil {
//								return err
//							}
//						} // TODO: Handle case when appUser.GoogleUniqueUserID is not empty
//					} else {
//						oldUser := models.AppUser{ID: entity.AppUserIntID}
//						newUser := models.AppUser{ID: appUserID}
//
//						if err = dal.DB.GetMulti(c, []db.EntityHolder{&oldUser, &newUser}); err != nil {
//							return
//						}
//
//						oldUser.GoogleUniqueUserID = ""
//						newUser.GoogleUniqueUserID = user.ID
//
//						if err = dal.DB.UpdateMulti(c, []db.EntityHolder{&oldUser, &newUser}); err != nil {
//							return
//						}
//					}
//					entity.AppUserIntID = appUserID
//				}
//
//				if _, err = gaedb.Put(tc, key, entity); err != nil {
//					err = errors.Wrap(err, "Failed to save google user")
//					return
//				}
//			}
//			return
//		} else  if err != datastore.ErrNoSuchEntity {
//			err = errors.Wrapf(err, "Failed to get google user entity by key=%v", key)
//			return
//		}
//
//		isNewGoogleUser = true
//		now := time.Now()
//		entity = &models.UserGoogleEntity{
//			LastSignIn: now,
//			User:       user,
//			OwnedByUser: user.OwnedByUser{
//				AppUserIntID: appUserID,
//				DtCreated: now,
//			},
//		}
//
//		if entity.AppUserIntID != 0 {
//			if user, err := dal.User.GetUserByID(c, entity.AppUserIntID); err != nil {
//				return err
//			} else if user.GoogleUniqueUserID != entity.ID {
//				if user.GoogleUniqueUserID != "" {
//					log.Warningf(c, "TODO: Handle case when connect with to user with different linked Google ID")
//				}
//				user.GoogleUniqueUserID = entity.ID
//				if err = dal.User.SaveUser(c, user); err != nil {
//					return err
//				}
//			}
//		} else {
//			emailLowCase := strings.ToLower(user.Email)
//			query := datastore.NewQuery(models.AppUserKind).Filter("EmailAddress = ", emailLowCase).Limit(2)
//			var (
//				appUserKeys []*datastore.Key
//				appUsers    []models.AppUserEntity
//			)
//			if appUserKeys, err = query.GetAll(c, &appUsers); err != nil {
//				err = errors.Wrap(err, "Failed to load users by email")
//				return
//			}
//			switch len(appUserKeys) {
//			case 1:
//				entity.AppUserIntID = appUserKeys[0].IntegerID()
//			case 0:
//				query = datastore.NewQuery(models.UserGoogleKind).Filter("Email =", user.Email).Limit(2)
//				var (
//					googleUserKeys []*datastore.Key
//					googleUsers    []models.UserGoogleEntity
//				)
//				if googleUserKeys, err = query.GetAll(c, &googleUsers); err != nil {
//					err = errors.Wrap(err, "Failed to load google users by email")
//					return
//				}
//				switch len(googleUserKeys) {
//				case 1:
//					panic("TODO: We need to handle situation when user changed email and that email was linked to another google account")
//				case 2:
//					err = fmt.Errorf("Found > 1 google users for email=%v, %v", user.Email, googleUserKeys)
//					return
//				}
//
//				isNewAppUser = true
//				appUserKey := datastore.NewIncompleteKey(tc, models.AppUserKind, nil)
//				if strings.Index(remoteAddr, ":") >= 0 {
//					remoteAddr = strings.Split(remoteAddr, ":")[0]
//				}
//				appUser := models.AppUserEntity{
//					GoogleUniqueUserID: user.ID,
//					DtCreated:          now,
//					LastUserAgent:      userAgent,
//					LastUserIpAddress:  remoteAddr,
//					ContactDetails: models.ContactDetails{
//						EmailContact: models.EmailContact{
//							EmailAddress:         emailLowCase,
//							EmailAddressOriginal: user.Email,
//							EmailConfirmed:       true,
//						},
//					},
//				}
//				if appUserKey, err = gaedb.Put(tc, appUserKey, &appUser); err != nil {
//					err = errors.Wrap(err, "Failed to save app use entity")
//					return
//				}
//				entity.AppUserIntID = appUserKey.IntegerID()
//			default: // len(appUserKeys) > 1
//				err = fmt.Errorf("Found > 1 users for email=%v, %v", emailLowCase, appUserKeys)
//				return
//			}
//		}
//
//		if _, err = gaedb.Put(tc, key, entity); err != nil {
//			err = errors.Wrap(err, "Failed to save google use entity")
//			return
//		}
//		return
//	}, dal.CrossGroupTransaction)
//	return
//}
