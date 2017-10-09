package models

import (
	gaeuser "google.golang.org/appengine/user" // TODO: Get rid of dependency to GAE?
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/pkg/errors"
	"github.com/strongo/app/db"
	"github.com/strongo/app/user"
)

const UserGoogleKind = "UserGoogle"

type UserGoogleEntity struct {
	gaeuser.User // TODO: We would wnat to abstract from a specific implementation
	user.Names
	user.LastLogin
	user.OwnedByUser
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

func (entity *UserGoogleEntity) Save() (properties []datastore.Property, err error) {
	if entity.AppUserIntID == 0 {
		err = errors.New("*UserGoogleEntity.Save() => AppUserIntID == 0")
		return
	}

	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}

	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"FederatedIdentity": gaedb.IsEmptyString,
		"FederatedProvider": gaedb.IsEmptyString,
		"AuthDomain":        gaedb.IsEmptyString,
		"ClientID":          gaedb.IsEmptyString,
		"ID":                gaedb.IsDuplicate,
		"AppUserID":         gaedb.IsZeroInt,
		"Admin":             gaedb.IsZeroBool,
	}); err != nil {
		return
	}

	for i, p := range properties {
		switch p.Name {
		case "FederatedIdentity":
		case "FederatedProvider":
		case "AuthDomain":
		case "ClientID":
		default:
			continue
		}
		p.NoIndex = true
		properties[i] = p
	}

	return
}

type UserGoogle struct {
	db.NoIntID
	ID string
	*UserGoogleEntity
}

var _ db.EntityHolder = (*UserGoogle)(nil)

func (userGoogle UserGoogle) UserAccount() user.Account {
	return user.Account{Provider: "google", ID: userGoogle.ID}
}

func (userGoogle UserGoogle) StrID() string {
	return userGoogle.ID
}

func (userGoogle *UserGoogle) SetStrID(id string) {
	userGoogle.ID = id
}

func (userGoogle UserGoogle) Kind() string {
	return UserGoogleKind
}

func (userGoogle *UserGoogle) SetEntity(v interface{}) {
	userGoogle.UserGoogleEntity = v.(*UserGoogleEntity)
}

func (userGoogle UserGoogle) Entity() interface{} {
	return userGoogle.UserGoogleEntity
}

//func (userGoogle *UserGoogle) SetStrID(id string) {
//	userGoogle.ID = id
//}
