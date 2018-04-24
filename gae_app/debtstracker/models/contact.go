package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"context"
	"google.golang.org/appengine/datastore"
)

func NewContactEntity(userID int64, details ContactDetails) *ContactEntity {
	return &ContactEntity{
		Status:         STATUS_ACTIVE,
		UserID:         userID,
		DtCreated:      time.Now(), // TODO: Should we pass from outside as parameter?
		ContactDetails: details,
	}
}

const ContactKind = "Counterparty" // TODO: Change value to Contact & migrated DB records

type Contact struct {
	db.IntegerID
	*ContactEntity
}

var _ db.EntityHolder = (*Contact)(nil)

func (Contact) Kind() string {
	return ContactKind
}

func (c *Contact) Entity() interface{} {
	return c.ContactEntity
}

func (Contact) NewEntity() interface{} {
	return new(ContactEntity)
}

func (c *Contact) SetEntity(entity interface{}) {
	if entity == nil {
		c.ContactEntity = nil
	} else {
		c.ContactEntity = entity.(*ContactEntity)
	}
}

func (c Contact) MustMatchCounterparty(counterparty Contact) {
	if !c.Balance().Equal(counterparty.Balance().Reversed()) {
		panic(fmt.Sprintf("contact[%d].Balance() != counterpartyContact[%d].Balance(): %v != %v", c.ID, counterparty.ID, c.Balance(), counterparty.Balance()))
	}
	if c.BalanceCount != counterparty.BalanceCount {
		panic(fmt.Sprintf("contact.BalanceCount != counterpartyContact.BalanceCount:  %v != %v", c.BalanceCount, counterparty.BalanceCount))
	}
}

func NewContact(id int64, entity *ContactEntity) Contact {
	return Contact{IntegerID: db.IntegerID{ID: id}, ContactEntity: entity}
}

type ContactEntity struct {
	DtCreated                  time.Time `datastore:",omitempty"`
	UserID                     int64     // owner can not be in parent key as we have problem with filtering transfers then
	CounterpartyUserID         int64     // The counterparty user ID if registered
	CounterpartyCounterpartyID int64
	LinkedBy                   string `datastore:",noindex"`
	//
	Status string
	ContactDetails
	Balanced
	TransfersJson string `datastore:",noindex"`
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

func (entity ContactEntity) GetTransfersInfo() (transfersInfo *UserContactTransfersInfo) {
	if entity.TransfersJson == "" {
		return &UserContactTransfersInfo{}
	}
	transfersInfo = new(UserContactTransfersInfo)
	if err := ffjson.Unmarshal([]byte(entity.TransfersJson), transfersInfo); err != nil {
		panic(err)
	}
	return
}

func (entity *ContactEntity) SetTransfersInfo(transfersInfo UserContactTransfersInfo) error {
	if data, err := ffjson.Marshal(transfersInfo); err != nil {
		return err
	} else {
		entity.TransfersJson = string(data)
		return nil
	}
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
	"TransfersJson":              gaedb.IsEmptyJson,
	"Nickname":                   gaedb.IsEmptyString,
	"FirstName":                  gaedb.IsEmptyString,
	"LastName":                   gaedb.IsEmptyString,
	"ScreenName":                 gaedb.IsEmptyString,
	"PhoneNumber":                gaedb.IsZeroInt,
	"PhoneNumberConfirmed":       gaedb.IsFalse,
	"EmailConfirmed":             gaedb.IsFalse,
	"TelegramUserID":             gaedb.IsZeroInt,
}

func (entity *ContactEntity) Save() (properties []datastore.Property, err error) {
	entity.EmailAddressOriginal = strings.TrimSpace(entity.EmailAddressOriginal)
	entity.EmailAddress = strings.ToLower(entity.EmailAddressOriginal)
	//entity.UpdateSearchName()

	if properties, err = datastore.SaveStruct(entity); err != nil {
		return
	}

	if properties, err = gaedb.CleanProperties(properties, contactPropertiesToClean); err != nil {
		return
	}

	checkHasProperties(ContactKind, properties)

	return
}

func (entity *ContactEntity) BalanceWithInterest(c context.Context, periodEnds time.Time) (balance Balance, err error) {
	balance = entity.Balance()
	if transferInfo := entity.GetTransfersInfo(); transferInfo != nil {
		err = updateBalanceWithInterest(true, balance, transferInfo.OutstandingWithInterest, periodEnds)
	}
	return
}

func ContactsByID(contacts []Contact) (contactsByID map[int64]*ContactEntity) {
	contactsByID = make(map[int64]*ContactEntity, len(contacts))
	for _, contact := range contacts {
		contactsByID[contact.ID] = contact.ContactEntity
	}
	return
}
