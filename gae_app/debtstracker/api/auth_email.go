package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/emails"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/user"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"golang.org/x/net/context"
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

	if _, err := dal.UserEmail.GetUserEmailByID(c, email); err != nil {
		if !db.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusInternalServerError, err)
			return
		}
	} else if err == nil {
		ErrorAsJson(c, w, http.StatusConflict, facade.ErrEmailAlreadyRegistered)
		return
	}

	if user, userEmail, err := facade.User.CreateUserByEmail(c, email, userName); err != nil {
		if errors.Cause(err) == facade.ErrEmailAlreadyRegistered {
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
		ReturnToken(c, w, user.ID, true, user.EmailAddress == "alexander.trakhimenok@gmail.com")
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

	userEmail, err := dal.UserEmail.GetUserEmailByID(c, email)
	if err != nil {
		if db.IsNotFound(err) {
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
	userEmail, err := dal.UserEmail.GetUserEmailByID(c, email)
	if db.IsNotFound(err) {
		ErrorAsJson(c, w, http.StatusForbidden, errors.New("Unknown email"))
		return
	}

	now := time.Now()

	pwdResetEntity := models.PasswordResetEntity{
		Email:  userEmail.ID,
		Status: "created",
		OwnedByUser: user.OwnedByUser{
			AppUserIntID: userEmail.AppUserIntID,
			DtCreated:    now,
		},
	}

	if _, err := dal.PasswordReset.CreatePasswordResetByID(c, &pwdResetEntity); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}

func handleChangePasswordAndSignIn(c context.Context, w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		passwordReset models.PasswordReset
	)

	if passwordReset.ID, err = strconv.ParseInt(r.PostFormValue("pin"), 10, 64); err != nil {
		ErrorAsJson(c, w, http.StatusBadRequest, err)
		return
	}

	pwd := r.PostFormValue("pwd")
	if pwd == "" {
		ErrorAsJson(c, w, http.StatusBadRequest, errors.New("Empty password"))
		return
	}

	if passwordReset, err = dal.PasswordReset.GetPasswordResetByID(c, passwordReset.ID); err != nil {
		if db.IsNotFound(err) {
			ErrorAsJson(c, w, http.StatusForbidden, errors.New("Unknown pin"))
			return
		}
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	isAdmin := IsAdmin(passwordReset.Email)

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {

		now := time.Now()
		user := models.AppUser{
			IntegerID:     db.NewIntID(passwordReset.AppUserIntID),
			AppUserEntity: new(models.AppUserEntity),
		}
		userEmail := models.UserEmail{
			StringID:        db.StringID{ID: models.GetEmailID(passwordReset.Email)},
			UserEmailEntity: new(models.UserEmailEntity),
		}

		entities := []db.EntityHolder{&user, &userEmail, &passwordReset}

		if err = dal.DB.GetMulti(c, entities); err != nil {
			return err
		}

		if err = userEmail.SetPassword(pwd); err != nil {
			return err
		}

		passwordReset.Status = "changed"
		passwordReset.Email = "" // Clean email as we don't need it anymore
		passwordReset.DtUpdated = now
		if changed := userEmail.AddProvider("password-reset"); changed {
			userEmail.DtUpdated = now
		}
		userEmail.SetLastLogin(now)
		user.SetLastLogin(now)

		if err = dal.DB.UpdateMulti(c, entities); err != nil {
			return err
		}
		return err
	}, dal.CrossGroupTransaction); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	ReturnToken(c, w, passwordReset.AppUserIntID, false, isAdmin)
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

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		now := time.Now()

		if userEmail, err = dal.UserEmail.GetUserEmailByID(c, userEmail.ID); err != nil {
			return err
		}

		var appUser models.AppUser
		if appUser, err = dal.User.GetUserByID(c, userEmail.AppUserIntID); err != nil {
			return err
		}

		if userEmail.ConfirmationPin() != pin {
			return errInvalidEmailConformationPin
		}

		userEmail.IsConfirmed = true
		userEmail.SetDtUpdated(now)
		userEmail.PasswordBcryptHash = []byte{}
		userEmail.SetLastLogin(now)
		appUser.SetLastLogin(now)

		entities := []db.EntityHolder{&appUser, &userEmail}
		if err = dal.DB.UpdateMulti(c, entities); err != nil {
			return err
		}
		return err
	}, dal.CrossGroupTransaction); err != nil {
		if db.IsNotFound(err) {
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
