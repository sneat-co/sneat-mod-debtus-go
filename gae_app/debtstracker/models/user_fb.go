package models

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app/user"
)

var _ user.AccountData = (*UserFacebookData)(nil)

var _ user.AccountRecord = (*UserFacebook)(nil)

type UserFacebook struct {
	// TODO: db.NoIntID - replace with DALGO
	record.WithID[string]
	FbAppOrPageID       string
	FbUserOrPageScopeID string
	data                *UserFacebookData
}

//func (u *UserFacebook) Key() *dal.Key {
//	return u.WithID.Key
//}

func (u *UserFacebook) Key() user.Account {
	return u.data.Account
}

func (u *UserFacebook) Data() user.AccountData {
	return u.data
}

func (u *UserFacebook) FbUserData() user.AccountData {
	return u.data
}

func (u *UserFacebook) Record() dal.Record {
	return u.WithID.Record
}

func (u *UserFacebook) AccountData() user.AccountData {
	return u.data
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
//	return UserFacebookCollection
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

// UserFacebookData - TODO: consider migrating to https://github.com/dal-go/dalgo4auth
type UserFacebookData struct {
	user.AccountDataBase
	Email            string `datastore:",noindex"`
	EmailIsConfirmed bool   `datastore:",noindex"`
	user.OwnedByUserWithID
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
