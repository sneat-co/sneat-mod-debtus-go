package models

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app/user"
)

const UserFacebookKind = "UserFb"

var _ user.AccountData = (*UserFacebookData)(nil)

var _ user.AccountRecord = (*UserFacebook)(nil)

type UserFacebook struct {
	// TODO: db.NoIntID - replace with DALGO
	record.WithID[string]
	FbAppOrPageID       string
	FbUserOrPageScopeID string
	Data                *UserFacebookData
}

func (u *UserFacebook) GetEmail() string {
	return u.Data.Email
}

func (u *UserFacebook) Key() *dal.Key {
	return u.WithID.Key
}

func (u *UserFacebook) Record() dal.Record {
	return u.WithID.Record
}

func (u *UserFacebook) AccountData() user.AccountData {
	return u.Data
}

//var _ user.AccountRecord = (*UserFacebook)(nil)

//var _ db.EntityHolder = (*UserFacebook)(nil)

func (u *UserFacebook) UserAccount() user.Account {
	return user.Account{Provider: "fb", App: u.FbAppOrPageID, ID: u.FbUserOrPageScopeID}
}

func UserFacebookID(fbAppOrPageID, fbUserOrPageScopeID string) string {
	return fbAppOrPageID + ":" + fbUserOrPageScopeID
}

//func (*UserFacebook) Kind() string {
//	return UserFacebookKind
//}

//func (UserFacebook) TypeOfID() db.TypeOfID {
//	return db.IsStringID
//}

func (u *UserFacebook) StrID() string {
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

//func (u *UserFacebook) Entity() interface{} {
//	return u.Data
//}
//
//func (UserFacebook) NewEntity() interface{} {
//	return new(UserFacebookData)
//}
//
//func (u *UserFacebook) SetEntity(entity interface{}) {
//	u.Data = entity.(*UserFacebookData)
//}

type UserFacebookData struct {
	user.LastLogin
	user.Names
	Email            string `datastore:",noindex"`
	EmailIsConfirmed bool   `datastore:",noindex"`
	user.OwnedByUserWithIntID
}

var _ user.AccountData = (*UserFacebookData)(nil)

func (entity UserFacebookData) GetEmail() string {
	return entity.Email
}

func (entity UserFacebookData) IsEmailConfirmed() bool {
	return entity.EmailIsConfirmed
}

//func (entity *UserFacebookData) Load(ps []datastore.Property) error {
//	if err := datastore.LoadStruct(entity, ps); err != nil {
//		return err
//	}
//	return nil
//}
//
//func (entity *UserFacebookData) Save() (properties []datastore.Property, err error) {
//	if err = entity.Validate(); err != nil {
//		return
//	}
//	if properties, err = datastore.SaveStruct(entity); err != nil {
//		err = errors.Wrap(err, "Failed to save struct to properties")
//		return
//	}
//	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
//		"FirsName": gaedb.IsEmptyString,
//		"LastName": gaedb.IsEmptyString,
//		"NickName": gaedb.IsEmptyString,
//	}); err != nil {
//		return
//	}
//	return
//}
