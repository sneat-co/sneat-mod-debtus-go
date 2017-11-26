package models

import (
	"github.com/pkg/errors"
	"github.com/strongo/app/user"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

const UserFacebookKind = "UserFb"

type UserFacebook struct {
	db.NoIntID
	FbAppOrPageID       string
	FbUserOrPageScopeID string
	*UserFacebookEntity
}

var _ user.AccountRecord = (*UserFacebook)(nil)

var _ db.EntityHolder = (*UserFacebook)(nil)

func (u UserFacebook) UserAccount() user.Account {
	return user.Account{Provider: "fb", App: u.FbAppOrPageID, ID: u.FbUserOrPageScopeID}
}

func UserFacebookID(fbAppOrPageID, fbUserOrPageScopeID string) string {
	return fbAppOrPageID + ":" + fbUserOrPageScopeID
}

func (UserFacebook) Kind() string {
	return UserFacebookKind
}

func (UserFacebook) TypeOfID() db.TypeOfID {
	return db.IsStringID
}

func (u UserFacebook) StrID() string {
	return UserFacebookID(u.FbAppOrPageID, u.FbUserOrPageScopeID)
}

func (u *UserFacebook) SetStrID(id string) {
	panic("Not implemented")
}

//func (u *UserFacebook) SetStrID(v string) {
//	vals := strings.Split(v, ":")
//	if len(vals) != 2 {
//		panic("Invalid id: " + v)
//	}
//	u.FbAppOrPageID = vals[0]
//	u.FbUserOrPageScopeID = vals[1]
//}

func (u *UserFacebook) Entity() interface{} {
	return u.UserFacebookEntity
}

func (UserFacebook) NewEntity() interface{} {
	return new(UserFacebookEntity)
}

func (u *UserFacebook) SetEntity(entity interface{}) {
	u.UserFacebookEntity = entity.(*UserFacebookEntity)
}

type UserFacebookEntity struct {
	user.LastLogin
	user.Names
	Email            string `datastore:",noindex"`
	EmailIsConfirmed bool   `datastore:",noindex"`
	user.OwnedByUser
}

var _ user.AccountEntity = (*UserFacebookEntity)(nil)

func (entity UserFacebookEntity) GetEmail() string {
	return entity.Email
}

func (entity UserFacebookEntity) IsEmailConfirmed() bool {
	return entity.EmailIsConfirmed
}

func (entity *UserFacebookEntity) Load(ps []datastore.Property) error {
	if err := datastore.LoadStruct(entity, ps); err != nil {
		return err
	}
	return nil
}

func (entity *UserFacebookEntity) Save() (properties []datastore.Property, err error) {
	if err = entity.Validate(); err != nil {
		return
	}
	if properties, err = datastore.SaveStruct(entity); err != nil {
		err = errors.Wrap(err, "Failed to save struct to properties")
		return
	}
	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"FirsName": gaedb.IsEmptyString,
		"LastName": gaedb.IsEmptyString,
		"NickName": gaedb.IsEmptyString,
	}); err != nil {
		return
	}
	return
}
