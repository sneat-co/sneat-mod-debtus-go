package models

import (
	"fmt"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/datastore"
	"strings"
)

func NewContactEntity(userID int64, details ContactDetails) *ContactEntity {
	return &ContactEntity{
		Status:         STATUS_ACTIVE,
		UserID:         userID,
		ContactDetails: details,
	}
}

const ContactKind = "Counterparty" // TODO: Change value to Contact & migrated DB records

type Contact struct {
	db.NoStrID
	ID int64
	*ContactEntity
}

func (Contact) Kind() string {
	return ContactKind
}

func (c Contact) IntID() int64 {
	return c.ID
}

func (c *Contact) Entity() interface{} {
	if c.ContactEntity == nil {
		c.ContactEntity = new(ContactEntity)
	}
	return c.ContactEntity
}

func (c *Contact) SetEntity(entity interface{}) {
	ce := entity.(*ContactEntity)
	c.ContactEntity = ce
}

func (c *Contact) SetIntID(id int64) {
	c.ID = id
}

func NewContact(id int64, entity *ContactEntity) Contact {
	return Contact{ID: id, ContactEntity: entity}
}

type ContactEntity struct {
	UserID                     int64 // owner can not be in parent key as we have problem with filtering transfers then
	CounterpartyUserID         int64 // The counterparty user ID if registered
	CounterpartyCounterpartyID int64
	//
	Status string
	ContactDetails
	Balanced
	SmsStats
	//
	//TelegramChatID int
	//
	//LastTransferID int64  `datastore:",noindex"` - Decided against as we do not need it really and would require either 2 Put() instead of 1 PutMulti()
	SearchName          []string `datastore:",noindex"` // Deprecated
	NoTransferUpdatesBy []string `datastore:",noindex"`
	GroupIDs            []string `datastore:",noindex"`
}

func (entity ContactEntity) String() string {
	return fmt.Sprintf("Contact{UserID: %v, CounterpartyUserID: %v, CounterpartyCounterpartyID: %v, Status: %v, ContactDetails: %v, Balance: '%v', LastTransferAt: %v}", entity.UserID, entity.CounterpartyUserID, entity.CounterpartyCounterpartyID, entity.Status, entity.ContactDetails, entity.BalanceJson, entity.LastTransferAt)
}

func (entity *ContactEntity) Info(counterpartyID int64, note, comment string) TransferCounterpartyInfo {
	return TransferCounterpartyInfo{
		ContactID:   counterpartyID,
		UserID:      entity.UserID,
		ContactName: entity.FullName(),
		Note:        note,
		Comment:     comment,
	}
}

//func (entity *ContactEntity) UpdateSearchName() {
//	fullName := entity.FullName()
//	entity.SearchName = []string{strings.ToLower(fullName)}
//	if entity.Username != "" {
//		username := strings.ToLower(fullName)
//		found := false
//		for _, searchName := range entity.SearchName {
//			if searchName == username {
//				found = true
//			}
//		}
//		if !found {
//			entity.SearchName = append(entity.SearchName, username)
//		}
//	}
//}

func (entity *ContactEntity) Load(ps []datastore.Property) error {
	p2 := make([]datastore.Property, 0, len(ps))
	for _, p := range ps {
		switch p.Name {
		case "SearchName": // Ignore legacy
		default:
			p2 = append(p2, p)
		}
	}
	if err := datastore.LoadStruct(entity, p2); err != nil {
		return err
	}
	if entity.PhoneNumberIsConfirmed && !entity.PhoneNumberConfirmed {
		entity.PhoneNumberConfirmed = true
	}
	return nil
}

func (entity *ContactEntity) Save() (properties []datastore.Property, err error) {
	entity.EmailAddressOriginal = strings.TrimSpace(entity.EmailAddressOriginal)
	entity.EmailAddress = strings.ToLower(entity.EmailAddressOriginal)
	//entity.UpdateSearchName()

	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}
	if err = checkHasProperties(AppUserKind, properties); err != nil {
		return
	}

	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		// Remove obsolete
		"PhoneNumberIsConfirmed": gaedb.IsObsolete,
		"SearchName":             gaedb.IsObsolete,
		// Remove defaults
		"CounterpartyUserID":         gaedb.IsZeroInt,
		"CounterpartyCounterpartyID": gaedb.IsZeroInt,
		"BalanceCount":               gaedb.IsZeroInt,
		"BalanceJson":                gaedb.IsEmptyStringOrSpecificValue("null"), //TODO: Remove once DB cleared
		"SmsCount":                   gaedb.IsZeroInt,
		"SmsCost":                    gaedb.IsZeroFloat,
		"SmsCostUSD":                 gaedb.IsZeroInt,
		"EmailAddress":               gaedb.IsEmptyString,
		"EmailAddressOriginal":       gaedb.IsEmptyString,
		"LastTransferID":             gaedb.IsZeroInt,
		"Nickname":                   gaedb.IsEmptyString,
		"FirstName":                  gaedb.IsEmptyString,
		"LastName":                   gaedb.IsEmptyString,
		"ScreenName":                 gaedb.IsEmptyString,
		"PhoneNumber":                gaedb.IsZeroInt,
		"PhoneNumberConfirmed":       gaedb.IsFalse,
		"EmailConfirmed":             gaedb.IsFalse,
		"TelegramUserID":             gaedb.IsZeroInt,
	}); err != nil {
		return
	}

	return
}

type test interface {
	datastore.PropertyLoadSaver
}
