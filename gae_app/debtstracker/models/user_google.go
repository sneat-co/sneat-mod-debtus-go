package models

import (
	"errors"
	"github.com/dal-go/dalgo/record"
	"github.com/strongo/app/user"
	"google.golang.org/appengine/datastore"
	gaeuser "google.golang.org/appengine/user" // TODO: Get rid of dependency to GAE?
)

const UserGoogleKind = "UserGoogle"

type UserGoogleEntity struct {
	gaeuser.User // TODO: We would want to abstract from a specific implementation
	user.Names
	user.LastLogin
	user.OwnedByUserWithIntID
}

var _ user.AccountEntity = (*UserGoogle)(nil)

func (entity UserGoogleEntity) GetEmail() string {
	return entity.Email
}

func (entity UserGoogleEntity) IsEmailConfirmed() bool {
	return entity.Email != ""
}

func (entity *UserGoogleEntity) Load(ps []datastore.Property) error {
	for i, p := range ps {
		if p.Name == "LastSignIn" {
			p.Name = "DtLastLogin"
			ps[i] = p
		}
	}
	return datastore.LoadStruct(entity, ps)
}

func (entity *UserGoogleEntity) Validate() (err error) {
	if entity.AppUserIntID == 0 {
		err = errors.New("*UserGoogleEntity.Save() => AppUserIntID == 0")
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

type UserGoogle struct { // TODO: Move out to library?
	record.WithID[string]
	*UserGoogleEntity
}

//var _ db.EntityHolder = (*UserGoogle)(nil)

func (userGoogle UserGoogle) UserAccount() user.Account {
	return user.Account{Provider: "google", App: "*", ID: userGoogle.ID}
}

func (userGoogle UserGoogle) Kind() string {
	return UserGoogleKind
}

func (userGoogle *UserGoogle) SetEntity(entity interface{}) {
	if entity == nil {
		userGoogle.UserGoogleEntity = nil
	} else {
		userGoogle.UserGoogleEntity = entity.(*UserGoogleEntity)
	}
}

func (userGoogle UserGoogle) Entity() interface{} {
	return userGoogle.UserGoogleEntity
}

func (UserGoogle) NewEntity() interface{} {
	return new(UserGoogleEntity)
}

//func (userGoogle *UserGoogle) SetStrID(id string) {
//	userGoogle.ID = id
//}
