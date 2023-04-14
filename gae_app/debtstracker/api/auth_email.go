package api

import (
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/emails"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/app/user"
	"github.com/strongo/log"
)

var (
	reEmail         = regexp.MustCompile(`.+@.+\.\w+`)
	ErrInvalidEmail = errors.New("Invalid email")
)

func validateEmail(email string) error {
	if !reEmail.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func handleSignUpWithEmail(c context.Context, w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.PostFormValue("email"))
	userName := strings.TrimSpace(r.PostFormValue("name"))

	if email == "" {
		BadRequestMessage(c, w, "Missing required value: email")
		return
	}

	if err := validateEmail(email); err != nil {
		ErrorAsJson(c, w, http.StatusBadRequest, err)
		return
	}

	if _, err := dtdal.UserEmail.GetUserEmailByID(c, nil, email); err != nil {
		if !dal.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
	} else if err == nil {
		ErrorAsJson(c, w, http.StatusConflict, facade.ErrEmailAlreadyRegistered)
		return
	}

	if user, userEmail, err := facade.User.CreateUserByEmail(c, email, userName); err != nil {
		if errors.Is(err, facade.ErrEmailAlreadyRegistered) {
			ErrorAsJson(c, w, http.StatusConflict, err)
			return
		} else {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
	} else {
		if err = emails.CreateConfirmationEmailAndQueueForSending(c, user, userEmail); err != nil {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
		ReturnToken(c, w, user.ID, true, user.Data.EmailAddress == "alexander.trakhimenok@gmail.com")
	}
}

func handleSignInWithEmail(c context.Context, w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.PostFormValue("email"))
	password := strings.TrimSpace(r.PostFormValue("password"))
	log.Debugf(c, "Email: %v", email)
	if email == "" || password == "" {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Missing required value"))
		return
	}

	if err := validateEmail(email); err != nil {
		jsonToResponse(c, w, map[string]string{"error": err.Error()})
		return
	}

	userEmail, err := dtdal.UserEmail.GetUserEmailByID(c, nil, email)
	if err != nil {
		if dal.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusForbidden, errors.New("Unknown email"))
		} else {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
		}
		return
	} else if err = userEmail.CheckPassword(password); err != nil {
		log.Debugf(c, "Invalid password: %v", err.Error())
		ErrorAsJson(c, w, http.StatusForbidden, errors.New("Invalid password"))
		return
	}

	ReturnToken(c, w, userEmail.AppUserIntID, false, userEmail.ID == "alexander.trakhimenok@gmail.com")
}

func handleRequestPasswordReset(c context.Context, w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	userEmail, err := dtdal.UserEmail.GetUserEmailByID(c, nil, email)
	if dal.IsNotFound(err) {
		ErrorAsJson(c, w, http.StatusForbidden, errors.New("Unknown email"))
		return
	}

	now := time.Now()

	pwdResetEntity := models.PasswordResetData{
		Email:                userEmail.ID,
		Status:               "created",
		OwnedByUserWithIntID: user.NewOwnedByUserWithIntID(userEmail.AppUserIntID, now),
	}

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
		_, err = dtdal.PasswordReset.CreatePasswordResetByID(c, tx, &pwdResetEntity)
		return err
	})
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	return
}

func handleChangePasswordAndSignIn(c context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		passwordReset models.PasswordReset
	)

	if passwordReset.ID, err = strconv.Atoi(r.PostFormValue("pin")); err != nil {
		ErrorAsJson(c, w, http.StatusBadRequest, err)
		return
	}

	pwd := r.PostFormValue("pwd")
	if pwd == "" {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Empty password"))
		return
	}

	if passwordReset, err = dtdal.PasswordReset.GetPasswordResetByID(c, nil, passwordReset.ID); err != nil {
		if dal.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusForbidden, errors.New("Unknown pin"))
			return
		}
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	isAdmin := IsAdmin(passwordReset.Data.Email)

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {

		now := time.Now()
		appUser := models.NewAppUser(passwordReset.Data.AppUserIntID, nil)
		userEmail := models.NewUserEmail(passwordReset.Data.Email, nil)

		records := []dal.Record{appUser.Record, userEmail.Record, passwordReset.Record}

		if err = dtdal.DB.GetMulti(c, records); err != nil {
			return err
		}

		if err = userEmail.SetPassword(pwd); err != nil {
			return err
		}

		passwordReset.Data.Status = "changed"
		passwordReset.Data.Email = "" // Clean email as we don't need it anymore
		passwordReset.Data.DtUpdated = now
		if changed := userEmail.AddProvider("password-reset"); changed {
			userEmail.DtUpdated = now
		}
		userEmail.SetLastLogin(now)
		appUser.Data.SetLastLogin(now)

		if err = tx.SetMulti(c, records); err != nil {
			return err
		}
		return err
	}); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	ReturnToken(c, w, passwordReset.Data.AppUserIntID, false, isAdmin)
}

var errInvalidEmailConformationPin = errors.New("email confirmation pin is not valid")

func handleConfirmEmailAndSignIn(c context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		userEmail models.UserEmail
		pin       string
	)

	userEmail.ID, pin = r.PostFormValue("email"), r.PostFormValue("pin")

	if userEmail.ID == "" {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Empty email"))
		return
	}
	if pin == "" {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Empty pin"))
		return
	}

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}
	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		now := time.Now()

		if userEmail, err = dtdal.UserEmail.GetUserEmailByID(c, tx, userEmail.ID); err != nil {
			return err
		}

		var appUser models.AppUser
		if appUser, err = facade.User.GetUserByID(c, tx, userEmail.AppUserIntID); err != nil {
			return err
		}

		if userEmail.ConfirmationPin() != pin {
			return errInvalidEmailConformationPin
		}

		userEmail.IsConfirmed = true
		userEmail.SetUpdatedTime(now)
		userEmail.PasswordBcryptHash = []byte{}
		userEmail.SetLastLogin(now)
		appUser.Data.SetLastLogin(now)

		entities := []dal.Record{appUser.Record, userEmail.Record}
		if err = tx.SetMulti(c, entities); err != nil {
			return err
		}
		return err
	}); err != nil {
		if dal.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusBadRequest, err)
			return
		} else if err == errInvalidEmailConformationPin {
			ErrorAsJson(c, w, http.StatusForbidden, err)
			return
		}
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	ReturnToken(c, w, userEmail.AppUserIntID, false, IsAdmin(userEmail.ID))
}
