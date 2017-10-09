package models

import (
	"github.com/pkg/errors"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/datastore"
	"time"
)

const EmailKind = "Email"

type Email struct {
	ID int64
	*EmailEntity
}

type EmailEntity struct {
	Status          string
	Error           string `datastore:",noindex"`
	DtCreated       time.Time
	DtSent          time.Time
	Subject         string `datastore:",noindex"`
	From            string `datastore:",noindex"`
	To              string
	BodyText        string `datastore:",noindex"`
	BodyHtml        string `datastore:",noindex"`
	AwsSesMessageID string
}

func (entity *EmailEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(entity, ps)
}

func (entity *EmailEntity) Save() (properties []datastore.Property, err error) {
	if entity.Status == "" {
		err = errors.New("email.Status is empty")
	}
	if entity.Subject == "" {
		err = errors.New("email.Subject is empty")
	}
	if entity.From == "" {
		err = errors.New("email.From is empty")
	}
	if entity.To == "" {
		err = errors.New("email.To is empty")
	}
	if entity.DtCreated.IsZero() {
		entity.DtCreated = time.Now()
	}
	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	return gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"DtSent":          gaedb.IsZeroTime,
		"AwsSesMessageID": gaedb.IsEmptyString,
		"Error":           gaedb.IsEmptyString,
		"BodyText":        gaedb.IsEmptyString,
		"BodyHtml":        gaedb.IsEmptyString,
	})
}
