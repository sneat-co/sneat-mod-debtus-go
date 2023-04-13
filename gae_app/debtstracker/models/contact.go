package models

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"reflect"
	"strings"
	"time"

	"context"
	"github.com/crediterra/money"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/v2/datastore"
)

func NewContactEntity(userID int64, details ContactDetails) *ContactData {
	return &ContactData{
		Status:         STATUS_ACTIVE,
		UserID:         userID,
		DtCreated:      time.Now(), // TODO: Should we pass from outside as parameter?
		ContactDetails: details,
	}
}

const ContactKind = "Counterparty" // TODO: Change value to Contact & migrated DB records

type Contact struct {
	record.WithID[int64]
	Data *ContactData
}

func NewContactKey(contactID int64) *dal.Key {
	if contactID == 0 {
		panic("NewContactKey(): contactID == 0")
	}
	return dal.NewKeyWithID(ContactKind, contactID)
}

func ContactRecords(contacts []Contact) (records []dal.Record) {
	records = make([]dal.Record, len(contacts))
	for i, contact := range contacts {
		records[i] = contact.Record
	}
	return
}

func NewContacts(ids ...int64) (contacts []Contact) {
	contacts = make([]Contact, len(ids))
	for i, id := range ids {
		if id == 0 {
			panic(fmt.Sprintf("ids[%d] == 0", i))
		}
		contacts[i] = NewContact(id, nil)
	}
	return
}

func NewContact(id int64, data *ContactData) Contact {
	key := NewContactKey(id)
	if data == nil {
		data = new(ContactData)
	}
	return Contact{
		WithID: record.NewWithID(id, key, data),
		Data:   data,
	}
}

func NewContactRecord() dal.Record {
	return dal.NewRecordWithIncompleteKey(ContactKind, reflect.Int64, new(ContactData))
}

//var _ db.EntityHolder = (*Contact)(nil)

//func (Contact) Kind() string {
//	return ContactKind
//}

//func (c *Contact) Entity() interface{} {
//	return c.Data
//}

//func (Contact) NewEntity() interface{} {
//	return new(ContactData)
//}

//func (c *Contact) SetEntity(entity interface{}) {
//	if entity == nil {
//		c.Data = nil
//	} else {
//		c.Data = entity.(*ContactData)
//	}
//}

func (c Contact) MustMatchCounterparty(counterparty Contact) {
	if !c.Data.Balance().Equal(counterparty.Data.Balance().Reversed()) {
		panic(fmt.Sprintf("contact[%d].Balance() != counterpartyContact[%d].Balance(): %v != %v", c.ID, counterparty.ID, c.Data.Balance(), counterparty.Data.Balance()))
	}
	if c.Data.BalanceCount != counterparty.Data.BalanceCount {
		panic(fmt.Sprintf("contact.BalanceCount != counterpartyContact.BalanceCount:  %v != %v", c.Data.BalanceCount, counterparty.Data.BalanceCount))
	}
}

type ContactData struct {
	DtCreated                  time.Time `datastore:",omitempty"`
	UserID                     int64     // owner can not be in parent key as we have problem with filtering transfers then
	CounterpartyUserID         int64     // The counterparty user ID if registered
	CounterpartyCounterpartyID int64
	LinkedBy                   string `datastore:",noindex"`
	//
	Status string
	ContactDetails
	money.Balanced
	TransfersJson string `datastore:",noindex"`
	SmsStats
	//
	//TelegramChatID int
	//
	//LasttransferID int  `datastore:",noindex"` - Decided against as we do not need it really and would require either 2 Put() instead of 1 PutMulti()
	SearchName          []string `datastore:",noindex"` // Deprecated
	NoTransferUpdatesBy []string `datastore:",noindex"`
	GroupIDs            []string `datastore:",noindex"`
}

func (entity *ContactData) String() string {
	return fmt.Sprintf("Contact{UserID: %v, CounterpartyUserID: %v, CounterpartyCounterpartyID: %v, Status: %v, ContactDetails: %v, Balance: '%v', LastTransferAt: %v}", entity.UserID, entity.CounterpartyUserID, entity.CounterpartyCounterpartyID, entity.Status, entity.ContactDetails, entity.BalanceJson, entity.LastTransferAt)
}

func (entity *ContactData) GetTransfersInfo() (transfersInfo *UserContactTransfersInfo) {
	if entity.TransfersJson == "" {
		return &UserContactTransfersInfo{}
	}
	transfersInfo = new(UserContactTransfersInfo)
	if err := ffjson.Unmarshal([]byte(entity.TransfersJson), transfersInfo); err != nil {
		panic(err)
	}
	return
}

func (entity *ContactData) SetTransfersInfo(transfersInfo UserContactTransfersInfo) error {
	if data, err := ffjson.Marshal(transfersInfo); err != nil {
		return err
	} else {
		entity.TransfersJson = string(data)
		return nil
	}
}

func (entity *ContactData) Info(counterpartyID int64, note, comment string) TransferCounterpartyInfo {
	return TransferCounterpartyInfo{
		ContactID:   counterpartyID,
		UserID:      entity.UserID,
		ContactName: entity.FullName(),
		Note:        note,
		Comment:     comment,
	}
}

//func (entity *ContactData) UpdateSearchName() {
//	fullName := entity.GetFullName()
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

func (entity *ContactData) Load(ps []datastore.Property) error {
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

var contactPropertiesToClean = map[string]gaedb.IsOkToRemove{
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
	"TransfersJson":              gaedb.IsEmptyJSON,
	"Nickname":                   gaedb.IsEmptyString,
	"FirstName":                  gaedb.IsEmptyString,
	"LastName":                   gaedb.IsEmptyString,
	"ScreenName":                 gaedb.IsEmptyString,
	"PhoneNumber":                gaedb.IsZeroInt,
	"PhoneNumberConfirmed":       gaedb.IsFalse,
	"EmailConfirmed":             gaedb.IsFalse,
	"TelegramUserID":             gaedb.IsZeroInt,
}

func (entity *ContactData) BeforeSave() (err error) {
	//entity.UpdateSearchName()
	entity.EmailAddressOriginal = strings.TrimSpace(entity.EmailAddressOriginal)
	entity.EmailAddress = strings.ToLower(entity.EmailAddressOriginal)
	return nil
}

func (entity *ContactData) Save() (properties []datastore.Property, err error) {
	if err = entity.BeforeSave(); err != nil {
		return
	}

	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}

	//if properties, err = gaedb.CleanProperties(properties, contactPropertiesToClean); err != nil {
	//	return
	//}

	//checkHasProperties(ContactKind, properties)

	return
}

func (entity *ContactData) BalanceWithInterest(c context.Context, periodEnds time.Time) (balance money.Balance, err error) {
	balance = entity.Balance()
	if transferInfo := entity.GetTransfersInfo(); transferInfo != nil {
		err = updateBalanceWithInterest(true, balance, transferInfo.OutstandingWithInterest, periodEnds)
	}
	return
}

func ContactsByID(contacts []Contact) (contactsByID map[int64]*ContactData) {
	contactsByID = make(map[int64]*ContactData, len(contacts))
	for _, contact := range contacts {
		contactsByID[contact.ID] = contact.Data
	}
	return
}
