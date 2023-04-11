package facade

import (
	"fmt"
	"math/rand"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
)

type authFacade struct {
}

var AuthFacade = authFacade{}

func (authFacade) AssignPinCode(c context.Context, loginID, userID int64) (loginPin models.LoginPin, err error) {
	err = dtdal.DB.RunInTransaction(c, func(c context.Context) error {
		if loginPin, err = dtdal.LoginPin.GetLoginPinByID(c, loginID); err != nil {
			return fmt.Errorf("failed to get LoginPin entity by ID=%s: %w", loginID, err)
		}
		if loginPin.UserID != 0 && loginPin.UserID != userID {
			return errors.New("LoginPin.UserID != userID")
		}
		if !loginPin.SignedIn.IsZero() {
			return errors.New("LoginPin.SignedIn.IsZero(): false")
		}
		random := rand.New(rand.NewSource(time.Now().UnixNano()))
		loginPin.Code = random.Int31n(9000) + 1000
		loginPin.UserID = userID
		loginPin.Pinned = time.Now()
		if err = dtdal.LoginPin.SaveLoginPin(c, loginPin); err != nil {
			return fmt.Errorf("failed to save LoginPin entity with ID=%v: %w", loginID, err)
		}
		return err
	}, nil)
	return
}

func (authFacade) SignInWithPin(c context.Context, loginID int64, loginPinCode int32) (userID int64, err error) {
	_ = loginPinCode
	var loginPin models.LoginPin
	return userID, dtdal.DB.RunInTransaction(c, func(c context.Context) error {
		if loginPin, err = dtdal.LoginPin.GetLoginPinByID(c, loginID); err != nil {
			return fmt.Errorf("failed to get LoginPin entity by ID=%v: %w", loginID, err)
		}
		if !loginPin.SignedIn.IsZero() {
			return ErrLoginAlreadySigned
		}
		if loginPin.Created.Add(time.Hour).Before(time.Now()) {
			return ErrLoginExpired
		}
		if userID = loginPin.UserID; userID == 0 {
			return errors.New("LoginPin.UserID == 0")
		}

		loginPin.SignedIn = time.Now()
		if err = dtdal.LoginPin.SaveLoginPin(c, loginPin); err != nil {
			return err
		}
		return err
	}, nil) // dtdal.CrossGroupTransaction)
}
