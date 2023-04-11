package gaedal

import (
	"github.com/crediterra/money"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/app/gae"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
	"google.golang.org/appengine/v2/delay"
)

func NewAppUserKey(c context.Context, appUserId int64) *datastore.Key {
	return gaedb.NewKey(c, models.AppUserKind, "", appUserId, nil)
}

func NewAppUserIncompleteKey(c context.Context) *datastore.Key {
	return gaedb.NewIncompleteKey(c, models.AppUserKind, nil)
}

type UserDalGae struct {
}

func NewUserDalGae() UserDalGae {
	return UserDalGae{}
}

var _ dtdal.UserDal = (*UserDalGae)(nil)

func (userDal UserDalGae) SetLastCurrency(c context.Context, userID int64, currency money.Currency) error {
	return dtdal.DB.RunInTransaction(c, func(c context.Context) error {
		user, err := facade.User.GetUserByID(c, userID)
		if err != nil {
			return err
		}
		user.SetLastCurrency(string(currency))
		return facade.User.SaveUser(c, user)
	}, dtdal.CrossGroupTransaction)

}

func (userDal UserDalGae) GetUserByStrID(c context.Context, userID string) (user models.AppUser, err error) {
	var intUserID int64
	if intUserID, err = strconv.ParseInt(userID, 10, 64); err != nil {
		err = fmt.Errorf("%w: UserDalGae.GetUserByStrID()", err)
		return
	}
	return facade.User.GetUserByID(c, intUserID)
}

func (userDal UserDalGae) GetUserByVkUserID(c context.Context, vkUserID int64) (models.AppUser, error) {
	query := datastore.NewQuery(models.AppUserKind).Filter("VkUserID =", vkUserID)
	return userDal.getUserByQuery(c, query, "VkUserID")
}

func (userDal UserDalGae) GetUserByEmail(c context.Context, email string) (models.AppUser, error) {
	email = strings.ToLower(email)
	query := datastore.NewQuery(models.AppUserKind).Filter("EmailAddress =", email).Filter("EmailConfirmed =", true).Limit(2)
	user, err := userDal.getUserByQuery(c, query, "EmailAddress, is confirmed")
	if user.ID == 0 && err == dal.ErrRecordNotFound {
		query = datastore.NewQuery(models.AppUserKind).Filter("EmailAddress =", email).Filter("EmailConfirmed =", false).Limit(2)
		user, err = userDal.getUserByQuery(c, query, "EmailAddress, is not confirmed")
	}
	log.Debugf(c, "GetUserByEmail() => err=%v, User(id=%d): %v", err, user.ID, user)
	return user, err
}

func (userDal UserDalGae) getUserByQuery(c context.Context, query *datastore.Query, searchCriteria string) (appUser models.AppUser, err error) {
	userEntities := make([]*models.AppUserEntity, 0, 2)
	var userKeys []*datastore.Key
	userKeys, err = query.GetAll(c, &userEntities)
	if err != nil {
		return
	}
	switch len(userKeys) {
	case 1:
		log.Debugf(c, "getUserByQuery(%v) => %v: %v", searchCriteria, userKeys[0].IntID(), userEntities[0])
		return models.AppUser{IntegerID: db.NewIntID(userKeys[0].IntID()), AppUserEntity: userEntities[0]}, nil
	case 0:
		err = dal.ErrRecordNotFound
		log.Debugf(c, "getUserByQuery(%v) => %v", searchCriteria, err)
		return
	default: // > 1
		errDup := db.ErrDuplicateUser{
			SearchCriteria:   searchCriteria,
			DuplicateUserIDs: make([]int64, len(userKeys)),
		}
		for i, userKey := range userKeys {
			errDup.DuplicateUserIDs[i] = userKey.IntID()
		}
		err = errDup
		return
	}
}

func (userDal UserDalGae) CreateAnonymousUser(c context.Context) (models.AppUser, error) {
	userKey := datastore.NewIncompleteKey(c, models.AppUserKind, nil)

	userEntity := models.AppUserEntity{
		IsAnonymous: true,
	}

	if userKey, err := gaedb.Put(c, userKey, &userEntity); err != nil {
		return models.AppUser{}, err
	} else {
		return models.AppUser{
			IntegerID:     db.NewIntID(userKey.IntID()),
			AppUserEntity: &userEntity,
		}, nil
	}
}

func (userDal UserDalGae) CreateUser(c context.Context, userEntity *models.AppUserEntity) (user models.AppUser, err error) {
	key := NewAppUserIncompleteKey(c)
	if key, err = gaedb.Put(c, key, userEntity); err != nil {
		return
	}
	user = models.AppUser{
		IntegerID:     db.NewIntID(key.IntID()),
		AppUserEntity: userEntity,
	}
	return
}

func (UserDalGae) DelayUpdateUserWithBill(c context.Context, userID, billID string) (err error) {
	if err = gae.CallDelayFunc(c, common.QUEUE_BILLS, "UpdateUserWithBill", delayedUpdateUserWithBill, userID, billID); err != nil {
		return
	}
	return
}

var delayedUpdateUserWithBill = delay.Func("delayedUpdateWithBill", func(c context.Context, userID, billID string) (err error) {
	var user models.AppUser

	if user, err = dtdal.User.GetUserByStrID(c, userID); err != nil {
		return
	}
	log.Debugf(c, "User: %v", user)
	return
})

func (UserDalGae) DelayUpdateUserWithContact(c context.Context, userID, billID int64) (err error) {
	if err = gae.CallDelayFuncWithDelay(c, time.Second/10, common.QUEUE_USERS, "updateUserWithContact", delayedUpdateUserWithContact, userID, billID); err != nil {
		return
	}
	return
}

var delayedUpdateUserWithContact = delay.Func("updateUserWithContact", updateUserWithContact)

func updateUserWithContact(c context.Context, userID, contactID int64) (err error) {
	log.Debugf(c, "updateUserWithContact(userID=%v, contactID=%v)", userID, contactID)
	var contact models.Contact
	if contact, err = facade.GetContactByID(c, contactID); err != nil {
		log.Errorf(c, "updateUserWithContact: %v", err)
		err = nil // TODO: Why ignore error here?
		return
	}
	return dtdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		var user models.AppUser

		if user, err = facade.User.GetUserByID(c, userID); err != nil {
			return
		}
		if dal.IsNotFound(err) {
			log.Errorf(c, err.Error())
			err = nil
		}

		if _, changed := user.AddOrUpdateContact(contact); changed {
			if err = facade.User.SaveUser(c, user); err != nil {
				return
			}
		} else {
			log.Debugf(c, "user not changed")
		}
		return
	}, nil)
}
