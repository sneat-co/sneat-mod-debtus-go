package facade

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/db"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/app/user"
	"github.com/strongo/log"
	gae_user "google.golang.org/appengine/user"
)

type userFacade struct {
}

var User = userFacade{}

var ErrEmailAlreadyRegistered = errors.New("Email already registered")

func (userFacade) GetUserByID(c context.Context, tx dal.ReadTransaction, userID int64) (user models.AppUser, err error) {
	key := dal.NewKeyWithID(models.AppUserKind, userID)
	user.Data = new(models.AppUserEntity)
	user.WithID = record.WithID[int64]{
		ID:     userID,
		Key:    key,
		Record: dal.NewRecordWithData(key, user.Data),
	}
	err = tx.Get(c, user.Record)
	return
}

func (userFacade) GetUsersByIDs(c context.Context, userIDs []int64) (users []*models.AppUser, err error) {
	//log.Debugf(c, "UserDalGae.GetUsersByIDs(%d)", userIDs)
	if len(userIDs) == 0 {
		return
	}
	entityHolders := db.CreateEntityHoldersWithIntIDs(userIDs, func() db.EntityHolder {
		return new(models.AppUser)
	})
	if err = dtdal.DB.GetMulti(c, entityHolders); err != nil {
		return
	}
	users = make([]*models.AppUser, len(entityHolders))
	for i, eh := range entityHolders {
		users[i] = eh.(*models.AppUser)
	}
	return
}

func (userFacade) SaveUser(c context.Context, tx dal.ReadwriteTransaction, user models.AppUser) (err error) {
	return tx.Set(c, user.Record)
}

func (uf userFacade) CreateUserByEmail(
	c context.Context,
	email, name string,
) (
	user models.AppUser,
	userEmail models.UserEmail,
	err error,
) {
	err = dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if userEmail, err = dtdal.UserEmail.GetUserEmailByID(c, email); err == nil {
			return ErrEmailAlreadyRegistered
		} else if !dal.IsNotFound(err) {
			return
		}

		if userEmail.ID == "" {
			log.Errorf(c, "userEmail.ID is empty string")
			userEmail.ID = strings.ToLower(strings.TrimSpace(email))
		}

		userEntity := dtdal.CreateUserEntity(dtdal.CreateUserData{
			ScreenName: name,
		})
		userEntity.AddAccount(userEmail.UserAccount())

		if user, err = dtdal.User.CreateUser(c, userEntity); err != nil {
			return
		}

		userEmail.UserEmailEntity = models.NewUserEmailEntity(user.ID, false, "email")
		if err = userEmail.SetPassword(dtdal.RandomCode(8)); err != nil {
			return
		}

		err = dtdal.UserEmail.SaveUserEmail(c, userEmail)
		return
	}, dtdal.CrossGroupTransaction)

	return
}

// This is used in invites.
func (uf userFacade) GetOrCreateEmailUser(
	c context.Context,
	email string,
	isConfirmed bool,
	createUserData *dtdal.CreateUserData,
	clientInfo models.ClientInfo,
) (
	userEmail models.UserEmail,
	isNewUser bool,
	err error,
) {

	var appUser models.AppUser

	if userEmail, err = dtdal.UserEmail.GetUserEmailByID(c, email); err == nil {
		return // User found
	} else if !dal.IsNotFound(err) { //
		return // Internal error
	}
	err = nil // Clear dtdal.ErrRecordNotFound

	now := time.Now()
	isNewUser = true
	userEmail = models.NewUserEmail(email, isConfirmed, "email")
	appUser = models.NewUser(models.ClientInfo{})
	appUser.Data.DtCreated = now
	appUser.Data.AddAccount(userEmail.UserAccount())

	var to db.RunOptions = dtdal.CrossGroupTransaction

	if err = dtdal.DB.RunInTransaction(c, func(tc context.Context) error {
		if err = User.SaveUser(tc, appUser); err != nil {
			return errors.Wrap(err, "Failed to save new appUser to datastore")
		}
		userEmail.DtCreated = now

		if err = dtdal.UserEmail.SaveUserEmail(c, userEmail); err != nil {
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
	if googleUser == nil {
		panic("googleUser == nil")
	}
	getUserAccountRecordFromDB := func(c context.Context) (user.AccountRecord, error) {
		userGoogle, err = dtdal.UserGoogle.GetUserGoogleByID(c, googleUser.ID)
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
				OwnedByUserWithIntID: user.OwnedByUserWithIntID{
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
	err = dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if userAccountRecord, err = getUserAccountRecordFromDB(c); err != nil && !dal.IsNotFound(err) {
			// Technical error
			return err
		}

		now := time.Now()

		isNewUser := userID == 0

		updateUser := func() {
			appUser.Data.SetLastLogin(now)
			appUser.Data.SetLastLogin(now)
			if !appUser.Data.EmailConfirmed && userAccountRecord.IsEmailConfirmed() {
				appUser.Data.EmailConfirmed = true
			}
			names := userAccountRecord.GetNames()
			if appUser.Data.FirstName == "" && names.FirstName != "" {
				appUser.Data.FirstName = names.FirstName
			}
			if appUser.Data.LastName == "" && names.LastName != "" {
				appUser.Data.LastName = names.LastName
			}
			if appUser.Data.Nickname == "" && names.NickName != "" {
				appUser.Data.Nickname = names.NickName
			}
		}

		if err == nil { // User account record found
			uaRecordUserID := userAccountRecord.GetAppUserID().(int64)
			if !isNewUser && uaRecordUserID != userID {
				panic(fmt.Sprintf("Relinking of appUser accounts us not implemented yet => userAccountRecord.GetAppUserIntID():%d != userID:%d", uaRecordUserID, userID))
			}
			if appUser, err = User.GetUserByID(c, uaRecordUserID); err != nil {
				if dal.IsNotFound(err) {
					err = fmt.Errorf("record UserGoogle is referencing non existing appUser: %w", err)
				}
				return
			}
			userAccountRecord.SetLastLogin(now)
			updateUser()

			if err = dtdal.DB.UpdateMulti(c, []db.EntityHolder{userAccountRecord, &appUser}); err != nil {
				return fmt.Errorf("failed to update User & UserFacebook with DtLastLogin: %w", err)
			}
			return
		}

		// UserGoogle record not found
		// Lets create new UserGoogle record
		if userAccountRecord, err = newUserAccountRecord(c); err != nil {
			return
		}

		if !isNewUser {
			if appUser, err = User.GetUserByID(c, tx, userID); err != nil {
				return
			}
		}

		if i, ok := userAccountRecord.(user.CreatedTimesSetter); ok {
			i.SetCreatedTime(now)
		}
		if i, ok := userAccountRecord.(user.UpdatedTimeSetter); ok {
			i.SetUpdatedTime(now)
		}
		userAccountRecord.SetLastLogin(now)

		email := models.GetEmailID(userAccountRecord.GetEmail())

		if email == "" {
			panic("Not implemented: userAccountRecord.GetEmail() returned empty string")
		}

		var userEmail models.UserEmail
		if userEmail, err = dtdal.UserEmail.GetUserEmailByID(c, email); err != nil && !dal.IsNotFound(err) {
			return // error
		}

		if dal.IsNotFound(err) { // UserEmail record NOT found
			userEmail := models.NewUserEmail(email, true, provider)
			userEmail.DtCreated = now

			// We need to create new User entity
			if isNewUser {
				appUser = models.NewUser(clientInfo)
				appUser.Data.DtCreated = now
			}
			appUser.Data.AddAccount(userAccountRecord.UserAccount()) // No need to check for changed as new appUser
			appUser.Data.AddAccount(userEmail.UserAccount())         // No need to check for changed as new appUser
			updateUser()

			if isNewUser {
				if appUser, err = dtdal.User.CreateUser(c, appUser.Data); err != nil {
					return
				}
			} else if err = User.SaveUser(c, tx, appUser); err != nil {
				return
			}

			userAccountRecord.(user.BelongsToUserWithIntID).SetAppUserIntID(appUser.ID)
			userEmail.AppUserIntID = appUser.ID
			if err = dtdal.DB.UpdateMulti(c, []db.EntityHolder{userAccountRecord, &userEmail}); err != nil {
				return
			}
			return
		} else { // UserEmail record found
			userAccountRecord.(user.BelongsToUserWithIntID).SetAppUserIntID(userEmail.AppUserIntID) // No need to create a new appUser, link to existing
			if !isNewUser && userEmail.AppUserIntID != userID {
				panic(fmt.Sprintf("Relinking of appUser accounts us not implemented yet => userEmail.AppUserIntID:%d != userID:%d", userEmail.AppUserIntID, userID))
			}

			if isNewUser {
				if appUser, err = User.GetUserByID(c, userEmail.AppUserIntID); err != nil {
					if dal.IsNotFound(err) {
						err = fmt.Errorf("record UserEmail is referencing non existing User: %w", err)
					}
					return
				}
			}

			if changed := userEmail.AddProvider(provider); changed || !userEmail.IsConfirmed {
				userEmail.IsConfirmed = true
				if err = dtdal.UserEmail.SaveUserEmail(c, userEmail); err != nil {
					return
				}
			}
			appUser.Data.AddAccount(userAccountRecord.UserAccount())
			updateUser()
			if err = dtdal.DB.UpdateMulti(c, []db.EntityHolder{userAccountRecord, &appUser}); err != nil {
				return fmt.Errorf("failed to create UserFacebook & update User: %w", err)
			}
			return
		}
	}, dtdal.CrossGroupTransaction)
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
		if userFacebook, err = dtdal.UserFacebook.GetFbUserByFbID(c, fbAppOrPageID, fbUserOrPageScopeID); err != nil {
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
				OwnedByUserWithIntID: user.OwnedByUserWithIntID{
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
