package models

import (
	"errors"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app/user"
	gaeuser "google.golang.org/appengine/user" // TODO: Get rid of dependency to GAE?
)

const UserGoogleKind = "UserGoogle"

type UserGoogleData struct {
	gaeuser.User // TODO: We would want to abstract from a specific implementation
	user.Names
	user.LastLogin
	user.OwnedByUserWithIntID
}

var _ user.AccountData = (*UserGoogleData)(nil)

func (entity *UserGoogleData) GetEmail() string {
	return entity.Email
}

func (entity *UserGoogleData) IsEmailConfirmed() bool {
	return entity.Email != ""
}

//func (entity *UserGoogleData) Load(ps []datastore.Property) error {
//	for i, p := range ps {
//		if p.Name == "LastSignIn" {
//			p.Name = "DtLastLogin"
//			ps[i] = p
//		}
//	}
//	return datastore.LoadStruct(entity, ps)
//}

func (entity *UserGoogleData) Validate() (err error) {
	if entity.AppUserIntID == 0 {
		err = errors.New("*UserGoogleData.Save() => AppUserIntID == 0")
		return
	}

	//if properties, err = datastore.SaveStruct(entity); err != nil {
	//	return
	//}
	//
	//if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
	//	"FederatedIdentity": gaedb.IsEmptyString,
	//	"FederatedProvider": gaedb.IsEmptyString,
	//	"AuthDomain":        gaedb.IsEmptyString,
	//	"ClientID":          gaedb.IsEmptyString,
	//	"ID":                gaedb.IsDuplicate,
	//	"AppUserID":         gaedb.IsZeroInt,
	//	"Admin":             gaedb.IsZeroBool,
	//}); err != nil {
	//	return
	//}
	//
	//for i, p := range properties {
	//	switch p.Name {
	//	case "FederatedIdentity":
	//	case "FederatedProvider":
	//	case "AuthDomain":
	//	case "ClientID":
	//	default:
	//		continue
	//	}
	//	p.NoIndex = true
	//	properties[i] = p
	//}

	return
}

var _ user.AccountRecord = (*UserGoogle)(nil)

type UserGoogle struct { // TODO: Move out to library?
	record.WithID[string]
	Data *UserGoogleData
}

func (userGoogle UserGoogle) Key() *dal.Key {
	return userGoogle.Key()
}

func (userGoogle UserGoogle) Record() dal.Record {
	return userGoogle.Record()
}

func (userGoogle UserGoogle) AccountData() user.AccountData {
	return userGoogle.Data
}

func (userGoogle UserGoogle) GetEmail() string {
	//TODO implement me
	panic("implement me")
}

func NewUserGoogle(id string) UserGoogle {
	key := dal.NewKeyWithID(UserGoogleKind, id)
	data := new(UserGoogleData)
	return UserGoogle{
		WithID: record.WithID[string]{
			ID:     id,
			Key:    key,
			Record: dal.NewRecordWithData(key, data),
		},
		Data: data,
	}
}

//var _ db.EntityHolder = (*UserGoogle)(nil)

func (userGoogle UserGoogle) UserAccount() user.Account {
	return user.Account{Provider: "google", App: "*", ID: userGoogle.ID}
}

//func (userGoogle UserGoogle) Kind() string {
//	return UserGoogleKind
//}
//
//func (userGoogle *UserGoogle) SetEntity(entity interface{}) {
//	if entity == nil {
//		userGoogle.UserGoogleData = nil
//	} else {
//		userGoogle.UserGoogleData = entity.(*UserGoogleData)
//	}
//}
//
//func (userGoogle UserGoogle) Entity() interface{} {
//	return userGoogle.UserGoogleData
//}

//func (UserGoogle) NewEntity() interface{} {
//	return new(UserGoogleData)
//}

//func (userGoogle *UserGoogle) SetStrID(id string) {
//	userGoogle.ID = id
//}
