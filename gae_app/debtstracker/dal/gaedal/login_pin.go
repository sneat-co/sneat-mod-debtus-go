package gaedal

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/db/gaedb"
)

type LoginPinDalGae struct {
}

func NewLoginPinDalGae() LoginPinDalGae {
	return LoginPinDalGae{}
}

func (LoginPinDalGae) GetLoginPinByID(c context.Context, id int64) (loginPin models.LoginPin, err error) {
	loginPin.ID = id
	entity := new(models.LoginPinEntity)
	if err = gaedb.Get(c, NewLoginPinKey(c, id), entity); err != nil {
		return
	}
	loginPin.LoginPinEntity = entity
	return
}

func (LoginPinDalGae) SaveLoginPin(c context.Context, loginPin models.LoginPin) (err error) {
	_, err = gaedb.Put(c, NewLoginPinKey(c, loginPin.ID), loginPin.LoginPinEntity)
	return
}

func (loginPinDalGae LoginPinDalGae) CreateLoginPin(c context.Context, channel, gaClientID string, createdUserID int64) (int64, error) {
	switch strings.ToLower(channel) {
	case "":
		return 0, errors.New("Parameter 'channel' is not set")
	case "telegram":
	case "viber":
	default:
		return 0, fmt.Errorf("Unknown channel: %v", channel)
	}
	if createdUserID != 0 {
		if _, err := facade.User.GetUserByID(c, createdUserID); err != nil {
			return 0, errors.Wrapf(err, "Unknown user ID: %d", createdUserID)
		}
	}

	entity := models.LoginPinEntity{
		Channel:    channel,
		Created:    time.Now(),
		UserID:     createdUserID,
		GaClientID: gaClientID,
	}
	if key, err := gaedb.Put(c, NewLoginPinIncompleteKey(c), &entity); err != nil {
		return 0, err
	} else {
		return key.IntID(), err
	}
}

//func (loginPinDalGae LoginPinDalGae) GetByID(c context.Context, loginID int64) (entity *models.LoginPinEntity, err error) {
//	entity = new(models.LoginPinEntity)
//	err = gaedb.Get(c, models.NewLoginPinKey(c, loginID), entity)
//	return
//}
