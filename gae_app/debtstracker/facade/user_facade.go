package facade

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/user"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
	gae_user "google.golang.org/appengine/user"
)

type userFacade struct {
}

var User = userFacade{}

var ErrEmailAlreadyRegistered = errors.New("Email already registered")

func (uf userFacade) CreateUserByEmail(
	c context.Context,
	email, name string,
) (
	user models.AppUser,
	userEmail models.UserEmail,
	err error,
) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if userEmail, err = dal.UserEmail.GetUserEmailByID(c, email); err == nil {
			return ErrEmailAlreadyRegistered
		} else if !db.IsNotFound(err) {
			return
		}

		if userEmail.ID == "" {
			log.Errorf(c, "userEmail.ID is empty string")
			userEmail.ID = strings.ToLower(strings.TrimSpace(email))
		}

		userEntity := dal.CreateUserEntity(dal.CreateUserData{
			ScreenName: name,
		})
		userEntity.AddAccount(userEmail.UserAccount())

		if user, err = dal.User.CreateUser(c, userEntity); err != nil {
			return
		}

		userEmail.UserEmailEntity = models.NewUserEmailEntity(user.ID, false, "email")
		if err = userEmail.SetPassword(dal.RandomCode(8)); err != nil {
			return
		}

		err = dal.UserEmail.SaveUserEmail(c, userEmail)
		return
	}, dal.CrossGroupTransaction)

	return
}

// This is used in invites.
func (uf userFacade) GetOrCreateEmailUser(
	c context.Context,
	email string,
	isConfirmed bool,
	createUserData *dal.CreateUserData,
	clientInfo models.ClientInfo,
) (
	userEmail models.UserEmail,
	isNewUser bool,
	err error,
) {

	var appUser models.AppUser

	if userEmail, err = dal.UserEmail.GetUserEmailByID(c, email); err == nil {
		return // User found
	} else if !db.IsNotFound(err) { //
		return // Internal error
	}
	err = nil // Clear dal.ErrRecordNotFound

	now := time.Now()
	isNewUser = true
	userEmail = models.NewUserEmail(email, isConfirmed, "email")
	appUser = models.NewUser(models.ClientInfo{})
	appUser.DtCreated = now
	appUser.AddAccount(userEmail.UserAccount())

	var to db.RunOptions = dal.CrossGroupTransaction

	if err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
		if err = dal.User.SaveUser(tc, appUser); err != nil {
			return errors.Wrap(err, "Failed to save new appUser to datastore")
		}
		userEmail.DtCreated = now

		if err = dal.UserEmail.SaveUserEmail(c, userEmail); err != nil {
			return err
		}
		return nil
	}, to); err != nil {
		return
	}
	return
}

func (uf userFacade) GetOrCreateUserGoogleOnSignIn(
	c context.Context, googleUser *gae_user.User, appUserID int64, clientInfo models.ClientInfo,
) (
	userGoogle models.UserGoogle, appUser models.AppUser, err error,
) {
	getUserAccountRecordFromDB := func(c context.Context) (user.AccountRecord, error) {
		userGoogle, err = dal.UserGoogle.GetUserGoogleByID(c, googleUser.ID)
		return &userGoogle, err
	}
	newUserAccountRecord := func(c context.Context) (user.AccountRecord, error) {
		if googleUser.Email == "" {
			return nil, errors.New("Not implemented yet: Google did not provided appUser email")
		}
		userGoogle = models.UserGoogle{
			StringID: db.StringID{ID: googleUser.ID},
			UserGoogleEntity: &models.UserGoogleEntity{
				User: *googleUser,
				OwnedByUser: user.OwnedByUser{
					AppUserIntID: appUserID,
				},
			},
		}
		return &userGoogle, nil
	}

	if appUser, err = getOrCreateUserAccountRecordOnSignIn(
		c,
		"google",
		appUserID,
		getUserAccountRecordFromDB,
		newUserAccountRecord,
		clientInfo,
	); err != nil {
		return
	}
	return
}

func getOrCreateUserAccountRecordOnSignIn(
	c context.Context,
	provider string,
	userID int64,
	getUserAccountRecordFromDB func(c context.Context) (user.AccountRecord, error),
	newUserAccountRecord func(c context.Context) (user.AccountRecord, error),
	clientInfo models.ClientInfo,
) (
	appUser models.AppUser, err error,
) {
	log.Debugf(c, "getOrCreateUserAccountRecordOnSignIn(provider=%v, userID=%d)", provider, userID)
	var userAccountRecord user.AccountRecord
	err = dal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if userAccountRecord, err = getUserAccountRecordFromDB(c); err != nil && !db.IsNotFound(err) {
			// Technical error
			return err
		}

		now := time.Now()

		isNewUser := userID == 0

		updateUser := func() {
			appUser.SetLastLogin(now)
			appUser.SetLastLogin(now)
			if !appUser.EmailConfirmed && userAccountRecord.IsEmailConfirmed() {
				appUser.EmailConfirmed = true
			}
			names := userAccountRecord.GetNames()
			if appUser.FirstName == "" && names.FirstName != "" {
				appUser.FirstName = names.FirstName
			}
			if appUser.LastName == "" && names.LastName != "" {
				appUser.LastName = names.LastName
			}
			if appUser.Nickname == "" && names.NickName != "" {
				appUser.Nickname = names.NickName
			}
		}

		if err == nil { // User account record found
			uaRecordUserID := userAccountRecord.GetAppUserIntID()
			if !isNewUser && uaRecordUserID != userID {
				panic(fmt.Sprintf("Relinking of appUser accounts us not implemented yet => userAccountRecord.GetAppUserIntID():%d != userID:%d", uaRecordUserID, userID))
			}
			if appUser, err = dal.User.GetUserByID(c, uaRecordUserID); err != nil {
				if db.IsNotFound(err) {
					err = errors.WithMessage(err, "UserGoogle is referencing non existing appUser")
				}
				return
			}
			userAccountRecord.SetLastLogin(now)
			updateUser()

			if err = dal.DB.UpdateMulti(c, []db.EntityHolder{userAccountRecord, &appUser}); err != nil {
				return errors.WithMessage(err, "Failed to update User & UserFacebook with DtLastLogin")
			}
			return
		}

		// UserGoogle record not found
		// Lets create new UserGoogle record
		if userAccountRecord, err = newUserAccountRecord(c); err != nil {
			return
		}

		if !isNewUser {
			if appUser, err = dal.User.GetUserByID(c, userID); err != nil {
				return
			}
		}

		userAccountRecord.SetDtCreated(now)
		userAccountRecord.SetLastLogin(now)

		email := models.GetEmailID(userAccountRecord.GetEmail())

		if email == "" {
			panic("Not implemented: userAccountRecord.GetEmail() returned empty string")
		}

		var userEmail models.UserEmail
		if userEmail, err = dal.UserEmail.GetUserEmailByID(c, email); err != nil && !db.IsNotFound(err) {
			return // error
		}

		if db.IsNotFound(err) { // UserEmail record NOT found
			userEmail := models.NewUserEmail(email, true, provider)
			userEmail.DtCreated = now

			// We need to create new User entity
			if isNewUser {
				appUser = models.NewUser(clientInfo)
				appUser.DtCreated = now
			}
			appUser.AddAccount(userAccountRecord.UserAccount()) // No need to check for changed as new appUser
			appUser.AddAccount(userEmail.UserAccount())         // No need to check for changed as new appUser
			updateUser()

			if isNewUser {
				if appUser, err = dal.User.CreateUser(c, appUser.AppUserEntity); err != nil {
					return
				}
			} else if err = dal.User.SaveUser(c, appUser); err != nil {
				return
			}

			userAccountRecord.SetAppUserIntID(appUser.ID)
			userEmail.AppUserIntID = appUser.ID
			if err = dal.DB.UpdateMulti(c, []db.EntityHolder{userAccountRecord, &userEmail}); err != nil {
				return
			}
			return
		} else { // UserEmail record found
			userAccountRecord.SetAppUserIntID(userEmail.AppUserIntID) // No need to create a new appUser, link to existing
			if !isNewUser && userEmail.AppUserIntID != userID {
				panic(fmt.Sprintf("Relinking of appUser accounts us not implemented yet => userEmail.AppUserIntID:%d != userID:%d", userEmail.AppUserIntID, userID))
			}

			if isNewUser {
				if appUser, err = dal.User.GetUserByID(c, userEmail.AppUserIntID); err != nil {
					if db.IsNotFound(err) {
						err = errors.WithMessage(err, "UserEmail is referencing non existing User")
					}
					return
				}
			}

			if changed := userEmail.AddProvider(provider); changed || !userEmail.IsConfirmed {
				userEmail.IsConfirmed = true
				if err = dal.UserEmail.SaveUserEmail(c, userEmail); err != nil {
					return
				}
			}
			appUser.AddAccount(userAccountRecord.UserAccount())
			updateUser()
			if err = dal.DB.UpdateMulti(c, []db.EntityHolder{userAccountRecord, &appUser}); err != nil {
				return errors.WithMessage(err, "Failed to create UserFacebook & update User")
			}
			return
		}
	}, dal.CrossGroupTransaction)
	return
}

func (uf userFacade) GetOrCreateUserFacebookOnSignIn(
	c context.Context,
	appUserID int64,
	fbAppOrPageID, fbUserOrPageScopeID, firstName, lastName string,
	email string, isEmailConfirmed bool,
	clientInfo models.ClientInfo,
) (
	userFacebook models.UserFacebook, appUser models.AppUser, err error,
) {
	log.Debugf(c, "GetOrCreateUserFacebookOnSignIn(firstName=%v, lastName=%v)", firstName, lastName)
	if fbAppOrPageID == "" {
		panic("fbAppOrPageID is empty string")
	}
	if fbAppOrPageID == "" {
		panic("fbUserOrPageScopeID is empty string")
	}

	updateNames := func(entity *models.UserFacebookEntity) {
		if firstName != "" && userFacebook.FirstName != firstName {
			userFacebook.FirstName = firstName
		}
		if lastName != "" && userFacebook.LastName != lastName {
			userFacebook.LastName = lastName
		}
	}

	getUserAccountRecordFromDB := func(c context.Context) (user.AccountRecord, error) {
		if userFacebook, err = dal.UserFacebook.GetFbUserByFbID(c, fbAppOrPageID, fbUserOrPageScopeID); err != nil {
			return &userFacebook, err
		}
		updateNames(userFacebook.UserFacebookEntity)
		return &userFacebook, err
	}

	newUserAccountRecord := func(c context.Context) (user.AccountRecord, error) {
		userFacebook = models.UserFacebook{
			FbAppOrPageID:       fbAppOrPageID,
			FbUserOrPageScopeID: fbUserOrPageScopeID,
			UserFacebookEntity: &models.UserFacebookEntity{
				Email: email,
				Names: user.Names{
					FirstName: firstName,
					LastName:  lastName,
				},
				EmailIsConfirmed: isEmailConfirmed,
				OwnedByUser: user.OwnedByUser{
					AppUserIntID: appUserID,
				},
			},
		}
		updateNames(userFacebook.UserFacebookEntity)
		return &userFacebook, nil
	}
	if appUser, err = getOrCreateUserAccountRecordOnSignIn(
		c,
		"fb",
		appUserID,
		getUserAccountRecordFromDB,
		newUserAccountRecord,
		clientInfo,
	); err != nil {
		return
	}
	return
}
