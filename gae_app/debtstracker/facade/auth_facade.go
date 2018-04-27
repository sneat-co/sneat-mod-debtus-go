package facade

import (
	"fmt"
	"math/rand"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
)

type authFacade struct {
}

var AuthFacade = authFacade{}

func (authFacade) AssignPinCode(c context.Context, loginID, userID int64) (loginPin models.LoginPin, err error) {
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if loginPin, err = dal.LoginPin.GetLoginPinByID(c, loginID); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("Failed to get LoginPin entity by ID: %v", loginID))
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
		if err = dal.LoginPin.SaveLoginPin(c, loginPin); err != nil {
			return errors.Wrapf(err, "Failed to save LoginPin entity with ID: %v", loginID)
		}
		return err
	}, nil)
	return
}

func (authFacade) SignInWithPin(c context.Context, loginID int64, loginPinCode int32) (userID int64, err error) {
	var loginPin models.LoginPin
	err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if loginPin, err = dal.LoginPin.GetLoginPinByID(c, loginID); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("Failed to get LoginPin entity by ID: %v", loginID))
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
		if err = dal.LoginPin.SaveLoginPin(c, loginPin); err != nil {
			return err
		}
		return err
	}, nil) // dal.CrossGroupTransaction)
	return
}
