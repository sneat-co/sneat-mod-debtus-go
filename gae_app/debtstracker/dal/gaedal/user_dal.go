package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gae"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"strings"
	"strconv"
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

var _ dal.UserDal = (*UserDalGae)(nil)

func (userDal UserDalGae) SetLastCurrency(c context.Context, userID int64, currency models.Currency) error {
	return dal.DB.RunInTransaction(c, func(c context.Context) error {
		user, err := userDal.GetUserByID(c, userID)
		if err != nil {
			return err
		}
		user.SetLastCurrency(string(currency))
		return userDal.SaveUser(c, user)
	}, dal.CrossGroupTransaction)

}

func (userDal UserDalGae) SaveUser(c context.Context, user models.AppUser) (err error) {
	var key *datastore.Key
	if user.ID == 0 {
		key = NewAppUserIncompleteKey(c)
	} else {
		key = NewAppUserKey(c, user.ID)
	}
	if key, err = gaedb.Put(c, key, user.AppUserEntity); err == nil && user.ID == 0 {
		user.ID = key.IntID()
	}
	return
}

func (userDal UserDalGae) GetUserByStrID(c context.Context, userID string) (user models.AppUser, err error) {
	var intUserID int64
	if intUserID, err = strconv.ParseInt(userID, 10, 64); err != nil {
		err = errors.WithMessage(err, "UserDalGae.GetUserByStrID()")
		return
	}
	return userDal.GetUserByID(c, intUserID)
}

func (userDal UserDalGae) GetUserByID(c context.Context, userID int64) (user models.AppUser, err error) {
	//log.Debugf(c, "UserDalGae.GetUserByID(%d)", userID)
	if userID == 0 {
		panic("GetUserByID(userID == 0)")
	}
	user.ID = userID

	var userEntity models.AppUserEntity

	if err = gaedb.Get(c, NewAppUserKey(c, userID), &userEntity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = db.NewErrNotFoundByIntID(models.AppUserKind, userID, nil)
			return
		} else {
			err = errors.Wrap(err, "Failed to get userEntity by id")
			return
		}
	}
	user = models.AppUser{IntegerID: db.NewIntID(userID), AppUserEntity: &userEntity}
	return
}

func (userDal UserDalGae) GetUsersByIDs(c context.Context, userIDs []int64) (users []models.AppUser, err error) {
	//log.Debugf(c, "UserDalGae.GetUsersByIDs(%d)", userIDs)
	if len(userIDs) == 0 {
		return
	}

	keys := make([]*datastore.Key, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = NewAppUserKey(c, userID)
	}

	err = gaedb.GetMulti(c, keys, &users)
	return
}

func (userDal UserDalGae) GetUserByVkUserID(c context.Context, vkUserID int64) (models.AppUser, error) {
	query := datastore.NewQuery(models.AppUserKind).Filter("VkUserID =", vkUserID)
	return userDal.getUserByQuery(c, query, "VkUserID")
}

func (userDal UserDalGae) GetUserByEmail(c context.Context, email string) (models.AppUser, error) {
	email = strings.ToLower(email)
	query := datastore.NewQuery(models.AppUserKind).Filter("EmailAddress =", email).Filter("EmailConfirmed =", true).Limit(2)
	user, err := userDal.getUserByQuery(c, query, "EmailAddress, is confirmed")
	if user.ID == 0 && err == db.ErrRecordNotFound {
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
		err = db.ErrRecordNotFound
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

	if user, err = dal.User.GetUserByStrID(c, userID); err != nil {
		return
	}
	log.Debugf(c, "User: %v", user)
	return
})
