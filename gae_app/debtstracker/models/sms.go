package models

import (
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/record"
	"reflect"
	"time"

	"github.com/strongo/decimal"
	"github.com/strongo/gotwilio"
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
	record.WithID[string]
	Data *TwilioSmsData
}

func NewTwilioSmsRecord() dal.Record {
	return dal.NewRecordWithIncompleteKey(TwilioSmsKind, reflect.String, new(TwilioSmsData))
}

func NewTwilioSmsFromRecord(r dal.Record) TwilioSms {
	key := r.Key()
	return TwilioSms{
		WithID: record.WithID[string]{
			ID:     key.ID.(string),
			Key:    key,
			Record: r,
		},
		Data: r.Data().(*TwilioSmsData),
	}
}

func NewTwilioSmsFromRecords(r []dal.Record) []TwilioSms {
	result := make([]TwilioSms, len(r))
	for i, v := range r {
		result[i] = NewTwilioSmsFromRecord(v)
	}
	return result
}

func NewTwilioSms(smsID string, data *TwilioSmsData) TwilioSms {
	key := dal.NewKeyWithID(TwilioSmsKind, smsID)
	if data == nil {
		data = new(TwilioSmsData)
	}
	return TwilioSms{
		WithID: record.WithID[string]{
			ID:     smsID,
			Key:    key,
			Record: dal.NewRecordWithData(key, data),
		},
		Data: data,
	}
}

//func (TwilioSms) Kind() string {
//	return AppUserKind
//}
//
//func (u *TwilioSms) Entity() interface{} {
//	if u.TwilioSmsData == nil {
//		u.TwilioSmsData = new(TwilioSmsData)
//	}
//	return u.TwilioSmsData
//}
//
//func (u *TwilioSms) SetEntity(entity interface{}) {
//	u.TwilioSmsData = entity.(*TwilioSmsData)
//}

type TwilioSmsData struct {
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
	Price    float32             `datastore:",noindex"` // TODO: Remove obsolete
	PriceUSD decimal.Decimal64p2 `datastore:",noindex"`
	//URL         string   `datastore:",noindex"`

	//
	CreatorTgChatID             int64
	CreatorTgSmsStatusMessageID int `datastore:",noindex"`
}

func NewTwilioSmsFromSmsResponse(userID int64, response *gotwilio.SmsResponse) TwilioSmsData {
	entity := TwilioSmsData{
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
