package models

import (
	"github.com/pkg/errors"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/bots-framework/platforms/telegram"
	"google.golang.org/appengine/datastore"
	"time"
	"github.com/strongo/app/db"
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

var _ datastore.PropertyLoadSaver = (*ReminderEntity)(nil)

type Reminder struct {
	db.NoStrID
	ID int64
	*ReminderEntity
}

func (r *Reminder) Load(ps []datastore.Property) error {
	panic("Not supported")
}

func (r *Reminder) Save() ([]datastore.Property, error) {
	panic("Not supported")
}

type ReminderEntity struct {
	ParentReminderID    int64
	IsAutomatic         bool `datastore:",noindex"`
	IsRescheduled       bool `datastore:",noindex"`
	TransferID          int64
	DtNext              time.Time
	DtScheduled         time.Time `datastore:",noindex"` // DtNext moves here once sent, can be used for stats & troubleshooting
	Locale              string    `datastore:",noindex"`
	ClosedByTransferIDs []int64   `datastore:",noindex"` // TODO: Why do we need list of IDs here?
	SentVia             string
	Status              string
	UserID              int64
	CounterpartyID      int64 // If this field != 0 then r is to a counterparty
	DtCreated           time.Time
	DtUpdated           time.Time `datastore:",noindex"`
	DtSent              time.Time
	DtUsed              time.Time `datastore:",noindex"` // When user clicks "Yes/no returned"
	DtViewed            time.Time `datastore:",noindex"`
	DtDiscarded         time.Time `datastore:",noindex"`
	BotID               string    `datastore:",noindex"`
	ChatIntID           int64     `datastore:",noindex"`
	MessageIntID        int64     `datastore:",noindex"`
	MessageStrID        string    `datastore:",noindex"`
	ErrDetails          string    `datastore:",noindex"`
}

func (r *ReminderEntity) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(r, ps)
}

func (r *ReminderEntity) Save() (properties []datastore.Property, err error) {
	if err = r.validate(); err != nil {
		return nil, err
	}
	r.DtUpdated = time.Now()
	if properties, err = datastore.SaveStruct(r); err != nil {
		return
	}

	properties, err = gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"DtDiscarded":      gaedb.IsZeroTime,
		"DtNext":           gaedb.IsZeroTime,
		"DtScheduled":      gaedb.IsZeroTime,
		"DtSent":           gaedb.IsZeroTime,
		"DtUsed":           gaedb.IsZeroTime,
		"DtViewed":         gaedb.IsZeroTime,
		"ErrDetails":       gaedb.IsEmptyString,
		"IsAutomatic":      gaedb.IsFalse,
		"IsRescheduled":    gaedb.IsFalse,
		"Locale":           gaedb.IsEmptyString,
		"MessageIntID":     gaedb.IsZeroInt,
		"MessageStrID":     gaedb.IsEmptyString,
		"ParentReminderID": gaedb.IsZeroInt,
		"SentVia":          gaedb.IsEmptyString,
	})

	return
}

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
		SentVia:     telegram_bot.TelegramPlatformID,
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
