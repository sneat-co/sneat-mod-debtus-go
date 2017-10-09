package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
	"strings"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/datastore"
	"encoding/base64"
	"github.com/strongo/app/db"
	"github.com/strongo/app/user"
)

const UserEmailKind = "UserEmail"

type UserEmailEntity struct {
	user.LastLogin
	user.Names
	user.OwnedByUser
	IsConfirmed        bool
	PasswordBcryptHash []byte `datastore:",noindex"`
	Providers          []string `datastore:",noindex"` // E.g. facebook, vk, user
}

var _ user.AccountEntity = (*UserEmailEntity)(nil)

func (entity UserEmailEntity) ConfirmationPin() string {
	pin := base64.RawURLEncoding.EncodeToString(entity.PasswordBcryptHash)
	//if len(pin) > 20 {
	//	pin = pin[:20]
	//}
	return pin
}

type UserEmail struct {
	ID string
	user.Names
	db.NoIntID
	*UserEmailEntity
}

var _ user.AccountRecord = (*UserEmail)(nil)

func (userEmail UserEmail) UserAccount() user.Account {
	return user.Account{Provider: "email", ID: userEmail.ID}
}

func (userEmail UserEmail) StrID() string {
	return userEmail.ID
}

func (userEmail UserEmail) SetStrID(id string) {
	userEmail.ID = id
}

func (userEmail UserEmail) Kind() string {
	return UserEmailKind
}

func (userEmail *UserEmail) SetEntity(v interface{}) {
	userEmail.UserEmailEntity = v.(*UserEmailEntity)
}

func (userEmail UserEmail) Entity() interface{} {
	return userEmail.UserEmailEntity
}

//func (userEmail *UserEmail) SetStrID(id string) {
//	userEmail.ID = id
//}

func GetEmailID(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NewUserEmail(email string, isConfirmed bool, provider string) UserEmail {
	return UserEmail{
		ID:              GetEmailID(email),
		UserEmailEntity: NewUserEmailEntity(0, isConfirmed, provider),
	}
}

func (userEmail UserEmail) GetEmail() string {
	return userEmail.ID
}

func (entity *UserEmailEntity) IsEmailConfirmed() bool {
	return entity.IsConfirmed
}

func NewUserEmailEntity(userID int64, isConfirmed bool, provider string) *UserEmailEntity {
	entity := &UserEmailEntity{
		OwnedByUser: user.OwnedByUser{
			AppUserIntID: userID,
			DtCreated:    time.Now(),
		},
		IsConfirmed: isConfirmed,
	}
	entity.AddProvider(provider)
	return entity
}

const pwdSole = "85d80e53-"

func (entity *UserEmailEntity) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword(entity.PasswordBcryptHash, []byte(pwdSole+password))
}

func (entity *UserEmailEntity) SetPassword(password string) (err error) {
	entity.PasswordBcryptHash, err = bcrypt.GenerateFromPassword([]byte(pwdSole+password), 0)
	return
}

func (entity *UserEmailEntity) AddProvider(v string) (changed bool) {
	for _, p := range entity.Providers {
		if p == v {
			return
		}
	}
	entity.Providers = append(entity.Providers, v)
	changed = true
	return
}

func (entity *UserEmailEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *UserEmailEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	return gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"DtUpdated": gaedb.IsZeroTime,
		"FirstName": gaedb.IsEmptyString,
		"LastName":  gaedb.IsEmptyString,
		"NickName":  gaedb.IsEmptyString,
		"PasswordBcryptHash": gaedb.IsEmptyByteArray,
	})
}
