package models

import (
	"bitbucket.com/debtstracker/gae_app/general"
	"github.com/pkg/errors"
	"github.com/strongo/app/gaedb"
	"google.golang.org/appengine/datastore"
	"time"
	"github.com/strongo/app/db"
)

const (
	ReceiptKind = "Receipt"

	ReceiptStatusCreated      = "created"
	ReceiptStatusSent         = "sent"
	ReceiptStatusViewed       = "viewed"
	ReceiptStatusAcknowledged = "acknowledged"
)

var ReceiptStatuses = [4]string{
	ReceiptStatusCreated,
	ReceiptStatusSent,
	ReceiptStatusViewed,
	ReceiptStatusAcknowledged,
}

type Receipt struct {
	db.NoStrID
	ID int64
	*ReceiptEntity
}

func (_ *Receipt) Kind() string {
	return ReceiptKind
}

func (r *Receipt) IntID() int64 {
	return r.ID
}

func (r *Receipt) Entity() interface{} {
	if r.ReceiptEntity == nil {
		r.ReceiptEntity = new(ReceiptEntity)
	}
	return r.ReceiptEntity
}

func (r *Receipt) SetEntity(entity interface{}) {
	r.ReceiptEntity = entity.(*ReceiptEntity)
}

func (r *Receipt) SetIntID(id int64) {
	r.ID = id
}

func NewReceipt(id int64, entity *ReceiptEntity) Receipt {
	return Receipt{ID: id, ReceiptEntity: entity}
}

const (
	ReceiptForFrom = "from"
	ReceiptForTo   = "to"
)

type ReceiptFor string

type ReceiptEntity struct {
	Status               string
	TransferID           int64
	CreatorUserID        int64      // IMPORTANT: Can be different from transfer.CreatorUserID (usually same). Think of 3d party bills
	For                  ReceiptFor `datastore:",noindex"` // TODO: always fill. If receipt.CreatorUserID != transfer.CreatorUserID then receipt.For must be set to either "from" or "to"
	ViewedByUserIDs      []int64
	CounterpartyUserID   int64 // TODO: Is it always equal to AcknowledgedByUserID?
	AcknowledgedByUserID int64 // TODO: Is it always equal to CounterpartyUserID?
	general.CreatedOn
	DtCreated      time.Time
	DtSent         time.Time
	DtFailed       time.Time
	DtViewed       time.Time
	DtAcknowledged time.Time
	SentVia        string
	SentTo         string
	Lang           string `datastore:",noindex"`
	Error          string `datastore:",noindex"` //TODO: Need a comment on when it is used
}

func (receiptEntity ReceiptEntity) Validate() (err error) {
	if receiptEntity.TransferID == 0 {
		return errors.New("receipt.TransferID == 0")
	}
	if err = validateString("Unknown receipt.Status", receiptEntity.Status, ReceiptStatuses[:]); err != nil {
		return err
	}
	return nil
}

func NewReceiptEntity(creatorUserID, transferID, counterpartyUserID int64, lang, sentVia, sentTo string, createdOn general.CreatedOn) ReceiptEntity {
	return ReceiptEntity{
		CreatorUserID:      creatorUserID,
		CounterpartyUserID: counterpartyUserID,
		TransferID:         transferID,
		CreatedOn:          createdOn,
		DtCreated:          time.Now(),
		Lang:               lang,
		SentVia:            sentVia,
		SentTo:             sentTo,
		Status:             ReceiptStatusCreated,
	}
}

func (r *ReceiptEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(r, ps)
}

func (r *ReceiptEntity) Save() (properties []datastore.Property, err error) {
	if properties, err = datastore.SaveStruct(r); err != nil {
		return
	}

	properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"AcknowledgedByUserID	": gaedb.IsZeroInt,
		"DtAcknowledged": gaedb.IsZeroTime,
		"DtFailed":       gaedb.IsZeroTime,
		"DtSent":         gaedb.IsZeroTime,
		"DtViewed":       gaedb.IsZeroTime,
		"Error":          gaedb.IsEmptyString,
		"For":            gaedb.IsEmptyString,
		"SentTo":         gaedb.IsEmptyString,
		"SentVia":        gaedb.IsEmptyString,
	})

	return
}
