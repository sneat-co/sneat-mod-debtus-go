package models

import (
	"encoding/base64"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"strings"
	"time"

	"github.com/strongo/app/user"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine/v2/datastore"
)

const UserEmailKind = "UserEmail"

type UserEmailData struct {
	user.LastLogin
	user.Names
	user.OwnedByUserWithIntID
	IsConfirmed        bool
	PasswordBcryptHash []byte   `datastore:",noindex"`
	Providers          []string `datastore:",noindex"` // E.g. facebook, vk, user
}

var _ user.AccountData = (*UserEmailData)(nil)

func NewUserEmailKey(email string) *dal.Key {
	return dal.NewKeyWithID(UserEmailKind, GetEmailID(email))
}

func NewUserEmail(email string, data *UserEmailData) UserEmail {
	id := GetEmailID(email)
	if data == nil {
		data = new(UserEmailData)
	}
	return UserEmail{
		WithID:        record.NewWithID(id, NewUserEmailKey(email), data),
		UserEmailData: data,
	}
}

func (entity UserEmailData) ConfirmationPin() string {
	pin := base64.RawURLEncoding.EncodeToString(entity.PasswordBcryptHash)
	//if len(pin) > 20 {
	//	pin = pin[:20]
	//}
	return pin
}

type UserEmail struct {
	record.WithID[string]
	user.Names
	*UserEmailData
}

//var _ user.AccountRecord = (*UserEmail)(nil)

func (userEmail UserEmail) UserAccount() user.Account {
	return user.Account{Provider: "email", ID: userEmail.ID}
}

func (userEmail UserEmail) Kind() string {
	return UserEmailKind
}

func (userEmail *UserEmail) SetEntity(entity interface{}) {
	if entity == nil {
		userEmail.UserEmailData = entity.(*UserEmailData)
	} else {
		userEmail.UserEmailData = entity.(*UserEmailData)
	}
}

func (userEmail UserEmail) Entity() interface{} {
	return userEmail.UserEmailData
}

func (UserEmail) NewEntity() interface{} {
	return new(UserEmailData)
}

//func (userEmail *UserEmail) SetStrID(id string) {
//	userEmail.ID = id
//}

func GetEmailID(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

//func NewUserEmail(email string, isConfirmed bool, provider string) UserEmail {
//	return UserEmail{
//		WithID:        record.WithID[string]{ID: GetEmailID(email)},
//		UserEmailData: NewUserEmailData(0, isConfirmed, provider),
//	}
//}

func (userEmail UserEmail) GetEmail() string {
	return userEmail.ID
}

func (entity *UserEmailData) IsEmailConfirmed() bool {
	return entity.IsConfirmed
}

func NewUserEmailData(userID int64, isConfirmed bool, provider string) *UserEmailData {
	entity := &UserEmailData{
		OwnedByUserWithIntID: user.NewOwnedByUserWithIntID(userID, time.Now()),
		IsConfirmed:          isConfirmed,
	}
	entity.AddProvider(provider)
	return entity
}

const pwdSole = "85d80e53-"

func (entity *UserEmailData) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword(entity.PasswordBcryptHash, []byte(pwdSole+password))
}

func (entity *UserEmailData) SetPassword(password string) (err error) {
	entity.PasswordBcryptHash, err = bcrypt.GenerateFromPassword([]byte(pwdSole+password), 0)
	return
}

func (entity *UserEmailData) AddProvider(v string) (changed bool) {
	for _, p := range entity.Providers {
		if p == v {
			return
		}
	}
	entity.Providers = append(entity.Providers, v)
	changed = true
	return
}

func (entity *UserEmailData) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *UserEmailData) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	//return gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
	//	"DtUpdated":          gaedb.IsZeroTime,
	//	"FirstName":          gaedb.IsEmptyString,
	//	"LastName":           gaedb.IsEmptyString,
	//	"NickName":           gaedb.IsEmptyString,
	//	"PasswordBcryptHash": gaedb.IsEmptyByteArray,
	//})
	return nil, nil
}
