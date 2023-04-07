package models

import (
	"errors"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/dalgo/record"
	"google.golang.org/appengine/datastore"
	"time"
)

const (
	ReminderStatusCreated           = "created"
	ReminderStatusSending           = "sending"
	ReminderStatusFailed            = "failed"
	ReminderStatusSent              = "sent"
	ReminderStatusViewed            = "viewed"
	ReminderStatusRescheduled       = "rescheduled"
	ReminderStatusUsed              = "used"
	ReminderStatusDiscarded         = "discarded"
	ReminderStatusInvalidNoTransfer = "invalid:no-transfer"
)

var ReminderStatuses = []string{
	ReminderStatusCreated,
	ReminderStatusSending,
	ReminderStatusFailed,
	ReminderStatusSent,
	ReminderStatusViewed,
	ReminderStatusRescheduled,
	ReminderStatusUsed,
	ReminderStatusDiscarded,
	ReminderStatusInvalidNoTransfer,
}

const ReminderKind = "Reminder"

//var _ datastore.PropertyLoadSaver = (*ReminderEntity)(nil)

type Reminder struct {
	record.WithID[int]
	*ReminderEntity
}

//var _ db.EntityHolder = (*Reminder)(nil)

func NewReminder(id int, entity *ReminderEntity) Reminder {
	return Reminder{WithID: record.WithID[int]{ID: id}, ReminderEntity: entity}
}

func (Reminder) Kind() string {
	return ReminderKind
}

func (r Reminder) Entity() interface{} {
	return r.ReminderEntity
}
func (Reminder) NewEntity() interface{} {
	return new(ReminderEntity)
}
func (r *Reminder) SetEntity(entity interface{}) {
	r.ReminderEntity = entity.(*ReminderEntity)
}

func (r *Reminder) Load(ps []datastore.Property) error {
	panic("Not supported")
}

func (r *Reminder) Save() ([]datastore.Property, error) {
	panic("Not supported")
}

type ReminderEntity struct {
	ParentReminderID    int64 `datastore:",omitempty"`
	IsAutomatic         bool  `datastore:",noindex,omitempty"`
	IsRescheduled       bool  `datastore:",noindex,omitempty"`
	TransferID          int64
	DtNext              time.Time
	DtScheduled         time.Time `datastore:",noindex,omitempty"` // DtNext moves here once sent, can be used for stats & troubleshooting
	Locale              string    `datastore:",noindex"`
	ClosedByTransferIDs []int64   `datastore:",noindex"` // TODO: Why do we need list of IDs here?
	SentVia             string    `datastore:",omitempty"`
	Status              string
	UserID              int64
	CounterpartyID      int64 // If this field != 0 then r is to a counterparty
	DtCreated           time.Time
	DtUpdated           time.Time `datastore:",noindex,omitempty"`
	DtSent              time.Time `datastore:",omitempty"`
	DtUsed              time.Time `datastore:",noindex,omitempty"` // When user clicks "Yes/no returned"
	DtViewed            time.Time `datastore:",noindex,omitempty"`
	DtDiscarded         time.Time `datastore:",noindex,omitempty"`
	BotID               string    `datastore:",noindex,omitempty"`
	ChatIntID           int64     `datastore:",noindex,omitempty"`
	MessageIntID        int64     `datastore:",noindex,omitempty"`
	MessageStrID        string    `datastore:",noindex,omitempty"`
	ErrDetails          string    `datastore:",noindex,omitempty"`
}

//func (r *ReminderEntity) Save() (properties []datastore.Property, err error) {
//	if err = r.validate(); err != nil {
//		return nil, err
//	}
//	r.DtUpdated = time.Now()
//	if properties, err = datastore.SaveStruct(r); err != nil {
//		return
//	}
//
//	if properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
//		"DtDiscarded":      gaedb.IsZeroTime,
//		"DtNext":           gaedb.IsZeroTime,
//		"DtScheduled":      gaedb.IsZeroTime,
//		"DtSent":           gaedb.IsZeroTime,
//		"DtUsed":           gaedb.IsZeroTime,
//		"DtViewed":         gaedb.IsZeroTime,
//		"ErrDetails":       gaedb.IsEmptyString,
//		"IsAutomatic":      gaedb.IsFalse,
//		"IsRescheduled":    gaedb.IsFalse,
//		"Locale":           gaedb.IsEmptyString,
//		"MessageIntID":     gaedb.IsZeroInt,
//		"MessageStrID":     gaedb.IsEmptyString,
//		"ParentReminderID": gaedb.IsZeroInt,
//		"SentVia":          gaedb.IsEmptyString,
//	}); err != nil {
//		return
//	}
//
//	return
//}

func (r ReminderEntity) validate() (err error) {
	if err = validateString("Unknown reminder.Status", r.Status, ReminderStatuses); err != nil {
		return err
	}
	if r.TransferID == 0 {
		return errors.New("reminder.TransferID == 0")
	}
	if r.SentVia == "" {
		return errors.New("reminder.SentVia is empty")
	}
	if r.DtCreated.IsZero() {
		return errors.New("reminder.DtCreated.IsZero()")
	}
	if !r.DtSent.IsZero() && r.DtSent.Before(r.DtCreated) {
		return errors.New("reminder.DtSent.Before(n.DtCreated)")
	}
	if !r.DtViewed.IsZero() && r.DtViewed.Before(r.DtSent) {
		return errors.New("reminder.DtViewed.Before(n.DtSent)")
	}
	if r.ChatIntID != 0 && r.BotID == "" || r.ChatIntID == 0 && r.BotID != "" {
		return errors.New("r.TgChatID != 0 && r.TgBot == '' || r.TgChatID == 0 && r.TgBot != ''")
	}
	return nil
}

func NewReminderViaTelegram(botID string, chatID, userID, transferID int64, isAutomatic bool, next time.Time) (reminder ReminderEntity) {
	reminder = ReminderEntity{
		Status:      ReminderStatusCreated,
		SentVia:     telegram.PlatformID,
		BotID:       botID,
		ChatIntID:   chatID,
		UserID:      userID,
		TransferID:  transferID,
		DtCreated:   time.Now(),
		IsAutomatic: isAutomatic,
		DtNext:      next,
	}
	return
}

func (r *ReminderEntity) ScheduleNextReminder(parentReminderID int64, next time.Time) *ReminderEntity {
	reminder := *r
	reminder.ParentReminderID = parentReminderID
	reminder.Status = ReminderStatusRescheduled

	reminder.DtCreated = time.Now()
	reminder.DtNext = next
	reminder.Status = ReminderStatusCreated
	zero := time.Time{}
	reminder.DtSent = zero
	reminder.DtDiscarded = zero
	reminder.DtViewed = zero
	reminder.MessageStrID = ""
	reminder.MessageIntID = 0

	r.IsRescheduled = true
	return &reminder
}
