package models

import (
	"github.com/strongo/gotwilio"
	"time"
	"github.com/strongo/decimal"
	"github.com/strongo/app/db"
)

const SmsKind = "Sms"

type Sms struct {
	DtCreated  time.Time
	DtUpdate   time.Time
	DtSent     time.Time
	InviteCode string
	To         string
	From       string
	Status     string
	Price      float32 `datastore:",noindex"`
}

const TwilioSmsKind = "TwilioSms"

type TwilioSms struct {
	db.NoIntID
	ID string
	*TwilioSmsEntity
}

func (TwilioSms) Kind() string {
	return AppUserKind
}

func (u TwilioSms) StrID() string {
	return u.ID
}


func (u *TwilioSms) Entity() interface{} {
	if u.TwilioSmsEntity == nil {
		u.TwilioSmsEntity = new(TwilioSmsEntity)
	}
	return u.TwilioSmsEntity
}

func (u *TwilioSms) SetEntity(entity interface{}) {
	u.TwilioSmsEntity = entity.(*TwilioSmsEntity)
}


type TwilioSmsEntity struct {
	UserID      int64
	DtCreated   time.Time
	DtUpdated   time.Time
	DtDelivered time.Time
	DtSent      time.Time
	AccountSid  string `datastore:",noindex"`
	To          string
	From        string `datastore:",noindex"`
	MediaUrl    string `datastore:",noindex"`
	Body        string `datastore:",noindex"`
	Status      string
	Direction   string
	//ApiVersion  string   `datastore:",noindex"`
	Price float32 `datastore:",noindex"` // TODO: Remove obsolete
	PriceUSD decimal.Decimal64p2 `datastore:",noindex"`
	//Url         string   `datastore:",noindex"`

	//
	CreatorTgChatID             int64
	CreatorTgSmsStatusMessageID int `datastore:",noindex"`
}

func NewTwilioSmsFromSmsResponse(userID int64, response *gotwilio.SmsResponse) TwilioSmsEntity {
	entity := TwilioSmsEntity{
		UserID:     userID,
		DtCreated:  time.Now(),
		DtUpdated:  time.Now(),
		AccountSid: response.AccountSid,
		To:         response.To,
		From:       response.From,
		MediaUrl:   response.MediaUrl,
		Body:       response.Body,
		Status:     response.Status,
		Direction:  response.Direction,
	}
	if response.Price != nil {
		entity.Price = *response.Price
	}
	return entity
}
