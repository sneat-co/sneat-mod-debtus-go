package models

import (
	"fmt"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/general"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/decimal"
	"google.golang.org/appengine/datastore"
)

type TransferDirection string

func (d TransferDirection) Reverse() TransferDirection {
	switch d {
	case TransferDirectionUser2Counterparty:
		return TransferDirectionCounterparty2User
	case TransferDirectionCounterparty2User:
		return TransferDirectionUser2Counterparty
	default:
		panic("Reverse not supported for %v" + string(d))
	}
}

const ( // Transfer directions
	TransferDirectionUser2Counterparty = "u2c"
	TransferDirectionCounterparty2User = "c2u"
	TransferDirection3dParty           = "3d-party"
)

const ( // Transfer statuses
	TransferViewed   = "viewed"
	TransferAccepted = "accepted"
	TransferDeclined = "declined"
)

const TransferKind = "Transfer"

var _ datastore.PropertyLoadSaver = (*TransferEntity)(nil)

func NewTransfer(id int64, entity *TransferEntity) Transfer {
	if id == 0 {
		panic("id == 0")
	}
	if entity == nil {
		panic("entity == nil")
	}
	return Transfer{
		IntegerID:      db.IntegerID{ID: id},
		TransferEntity: entity,
	}
}

type Transfer struct {
	db.IntegerID
	*TransferEntity
}

var _ db.EntityHolder = (*Transfer)(nil)

func (Transfer) Kind() string {
	return TransferKind
}

func (t Transfer) IntID() int64 {
	return t.ID
}

func (t *Transfer) Entity() interface{} {
	return t.TransferEntity
}

func (Transfer) NewEntity() interface{} {
	return new(TransferEntity)
}

func (t *Transfer) SetEntity(entity interface{}) {
	if entity == nil {
		t.TransferEntity = nil
	} else {
		t.TransferEntity = entity.(*TransferEntity)
	}
}

func (t *Transfer) SetIntID(id int64) {
	t.ID = id
}

func (t *TransferEntity) HasObsoleteProps() bool {
	return t.hasObsoleteProps
}

func (t *TransferEntity) GetStartDate() time.Time {
	return t.DtCreated // TODO: Change to DtStart?
}

func (t *TransferEntity) GetLendingValue() decimal.Decimal64p2 {
	return t.AmountInCents
}

type TransferEntity struct {
	hasObsoleteProps bool
	general.CreatedOn
	from *TransferCounterpartyInfo
	to   *TransferCounterpartyInfo

	BillIDs []string

	SmsStats
	DirectionObsoleteProp string  `datastore:"Direction,noindex,omitempty"`
	IsReturn              bool    `datastore:",noindex,omitempty"` // We need it is not always possible to identify original transfer (think multiply & partial transfers)
	ReturnToTransferIDs   []int64 `datastore:",noindex"`           // List of transfer to which this debt is a return. Should be populated only if IsReturn=True
	//
	returns           []TransferReturnJson // Deserialized cache
	ReturnsJson       string               `datastore:",noindex,omitempty"`
	ReturnsCount      int                  `datastore:",noindex,omitempty"`
	ReturnTransferIDs []int64              `datastore:",noindex"` // TODO: Obsolete - replace with ReturnsJson List of transfers that return money to this debts
	//
	CreatorUserID           int64  `datastore:",noindex"`           // Do not delete
	CreatorCounterpartyID   int64  `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>ContactID
	CreatorCounterpartyName string `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>ContactName
	CreatorNote             string `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>Note
	CreatorComment          string `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>Comment

	CreatorTgReceiptByTgMsgID int64 `datastore:",noindex,omitempty"` // TODO: Move to Receipt ?
	//
	//CreatorTgBotID       string `datastore:",noindex"` // TODO: Migrated to TransferCounterpartyInfo
	//CreatorTgChatID      int64  `datastore:",noindex"` // TODO: Migrated to TransferCounterpartyInfo
	//CounterpartyTgBotID  string `datastore:",noindex"` // TODO: Migrated to TransferCounterpartyInfo
	//CounterpartyTgChatID int64  `datastore:",noindex"` // TODO: Migrated to TransferCounterpartyInfo
	//
	//CreatorAutoRemindersDisabled bool   `datastore:",noindex"`
	//CreatorReminderID      int64 `datastore:",noindex"` // obsolete
	//CounterpartyReminderID int64 `datastore:",noindex"` // obsolete
	//
	CounterpartyUserID           int64  `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>UserID
	CounterpartyCounterpartyID   int64  `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>ContactID
	CounterpartyCounterpartyName string `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>ContactName
	CounterpartyNote             string `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>Note
	CounterpartyComment          string `datastore:",noindex,omitempty"` //TODO: Replace with <From|To>Comment
	//CounterpartyAutoRemindersDisabled bool   `datastore:",noindex"`
	//CounterpartyTgReceiptInlineMessageID string    `datastore:",noindex"` - not useful as we can edit message just once on callback

	C_From string `datastore:",noindex"`
	C_To   string `datastore:",noindex"`

	//** New properties to replace Creator/Contact set of props **
	//FromUserID           int64  `datastore:",noindex"`
	//FromUserName         string `datastore:",noindex"`
	//FromCounterpartyID   int64  `datastore:",noindex"`
	//FromCounterpartyName string `datastore:",noindex"`
	//FromComment          string `datastore:",noindex"`
	//FromNote             string `datastore:",noindex"`
	//ToUserID             int64  `datastore:",noindex"`
	//ToUserName           string `datastore:",noindex"`
	//ToCounterpartyID     int64  `datastore:",noindex"`
	//ToCounterpartyName   string `datastore:",noindex"`
	//ToComment            string `datastore:",noindex"`
	//ToNote               string `datastore:",noindex"`

	AcknowledgeStatus string    `datastore:",noindex,omitempty"`
	AcknowledgeTime   time.Time `datastore:",noindex,omitempty"`

	// This 2 fields are used in conjunction with .Order("-DtCreated")
	BothUserIDs         []int64 // This is needed to show transactions by user regardless who created
	BothCounterpartyIDs []int64 // This is needed to show transactions by counterparty regardless who created
	//
	DtCreated time.Time
	DtDueOn   time.Time `datastore:",omitempty"`

	//Amount                   float64                                    // TODO: Obsolete!, Replaced with AmountInCents
	//AmountReturned           float64             `datastore:",noindex"` // TODO: Obsolete!, Replaced with AmountInCentsReturned
	//AmountOutstanding        float64             `datastore:",noindex"` // TODO: Obsolete!, Replaced with AmountInCentsOutstanding

	AmountInCents         decimal.Decimal64p2
	AmountInCentsReturned decimal.Decimal64p2 `datastore:",noindex,omitempty"`
	AmountInCentsInterest decimal.Decimal64p2 `datastore:",noindex,omitempty"`
	//AmountInCentsOutstanding decimal.Decimal64p2 `datastore:",noindex,omitempty"` // TODO: Should be removed!

	TransferInterest

	IsOutstanding bool
	Currency      Currency // Should be indexed for loading outstanding transfers
	//
	ReceiptsSentCount int64   `datastore:",noindex"`
	ReceiptIDs        []int64 `datastore:",noindex"`
}

func (t Transfer) String() string {
	if t.TransferEntity == nil {
		return fmt.Sprintf("Transfer{ID: %d, Entity: nil}", t.ID)
	} else {
		return fmt.Sprintf("Transfer{ID: %d, Entity: %v}", t.ID, t.TransferEntity)
	}
}

func (t TransferEntity) String() string {
	return fmt.Sprintf(
		"TransferEntity{DtCreated: %v, Direction: %v, GetAmount(): %v, AmoutInCentsReturned: %v, IsReturn: %v, ReturnToTransferIDs: %v, CreatorUserID: %d, Creator: %v, Contact: %v, BothUserIDs: %v, BothCounterpartyIDs: %v, From: %v, To: %v}",
		t.DtCreated, t.Direction(), t.GetAmount(), t.AmountInCentsReturned, t.IsReturn, t.ReturnToTransferIDs, t.CreatorUserID, t.Creator(), t.Counterparty(), t.BothUserIDs, t.BothCounterpartyIDs, t.From(), t.To())
}

func (t *TransferEntity) Direction() TransferDirection {
	if t.DirectionObsoleteProp != "" {
		return TransferDirection(t.DirectionObsoleteProp)
	}
	switch t.CreatorUserID {
	case 0:
		panic("CreatorUserID == 0")
	case t.From().UserID:
		return TransferDirectionUser2Counterparty
	case t.To().UserID:
		return TransferDirectionCounterparty2User
	}
	return TransferDirection3dParty
}

func (t *TransferEntity) DirectionForUser(userID int64) TransferDirection {
	switch userID {
	case t.From().UserID:
		return TransferDirectionUser2Counterparty
	case t.To().UserID:
		return TransferDirectionCounterparty2User
	case t.CreatorUserID:
		return TransferDirection3dParty
	default:
		panic(t.transferIsNotAssociatedWithUser(userID))
	}
}

func (t *TransferEntity) IsReverseDirection(t2 *TransferEntity) bool {
	return t.DirectionForUser(t.CreatorUserID) == t2.DirectionForUser(t.CreatorUserID).Reverse()
}

func (t *TransferEntity) DirectionForContact(contactID int64) TransferDirection {
	switch contactID {
	case t.From().ContactID:
		return TransferDirectionCounterparty2User
	case t.To().ContactID:
		return TransferDirectionUser2Counterparty
	default:
		panic(t.transferIsNotAssociatedWithContact(contactID))
	}
}

func (t *TransferEntity) transferIsNotAssociatedWithUser(userID int64) string {
	return fmt.Sprintf(
		"Transfer is not associated with userID=%d  (FromUserID=%d, ToUserID=%d)",
		userID, t.From().UserID, t.To().UserID,
	)
}

func (t *TransferEntity) transferIsNotAssociatedWithContact(contactID int64) string {
	return fmt.Sprintf(
		"Transfer is not associated with contactID=%v  (FromContactID=%v, ToContactID=%v)",
		contactID, t.From().ContactID, t.To().ContactID,
	)
}

func (t *TransferEntity) transferIsNotRelatedToCreator() string {
	return ErrTransferNotRelatedToCreator.Error() + fmt.Sprintf(
		"\nDirection(): %v, CreatorUserID: %d, From: %v, To: %v",
		t.Direction(), t.CreatorUserID, t.C_From, t.C_To,
	)
}

func (t *TransferEntity) ReturnDirectionForUser(userID int64) TransferDirection {
	switch userID {
	case 0:
		panic("userID == 0")
	case t.From().UserID:
		return TransferDirectionCounterparty2User
	case t.To().UserID:
		return TransferDirectionUser2Counterparty
	default:
		panic(t.transferIsNotAssociatedWithUser(userID))
	}
}

var ErrTransferNotRelatedToCreator = errors.New("Transfer is not related to creator")

func (t TransferEntity) Creator() *TransferCounterpartyInfo { // TODO: Same as t.Creator()
	if t.CreatorUserID == 0 {
		panic("CreatorUserID == 0")
	}
	if counterparty := t.From(); counterparty.UserID == t.CreatorUserID {
		return counterparty
	} else if counterparty = t.To(); counterparty.UserID == t.CreatorUserID {
		return counterparty
	}
	panic(t.transferIsNotRelatedToCreator())
}

func (t *TransferEntity) Counterparty() *TransferCounterpartyInfo {
	//return TransferCounterpartyInfo{
	//	UserID:         t.CounterpartyUserID,
	//	ContactID: t.CreatorCounterpartyID,
	//	ContactName:           t.CreatorCounterpartyName,
	//	Note:           t.CreatorNote,
	//	Comment:        t.CreatorComment,
	//}
	switch t.Direction() {
	case TransferDirectionUser2Counterparty:
		return t.To()
	case TransferDirectionCounterparty2User:
		return t.From()
	default:
		panic(t.transferIsNotRelatedToCreator())
	}
}

func (t TransferEntity) CounterpartyInfoByUserID(userID int64) *TransferCounterpartyInfo {
	switch userID {
	case t.From().UserID:
		return t.To()
	case t.To().UserID:
		return t.From()
	default:
		panic(t.transferIsNotAssociatedWithUser(userID))
	}
}

func (t TransferEntity) UserInfoByUserID(userID int64) *TransferCounterpartyInfo {
	switch userID {
	case t.From().UserID:
		return t.from
	case t.To().UserID:
		return t.to
	default:
		panic(t.transferIsNotAssociatedWithUser(userID))
	}
}

//const TRANSFER_REMINDERS_DISABLED = "disabled"
//
//func (t *Transfer) IsRemindersDisabled(userID int64) bool {
//	switch userID {
//	case t.CreatorUserID:
//		return t.CreatorAutoRemindersDisabled
//	case t.CounterpartyUserID:
//		return t.CounterpartyAutoRemindersDisabled
//	default:
//		panic("Attempt to check reminders for a user not related to the transfer")
//	}
//}
//
//// Returns true if value have been changed and false if unchanged.
//func (t *Transfer) setAutoRemindersDisabled(userID int64, value bool) bool {
//	switch userID {
//	case t.CreatorUserID:
//		if t.CreatorAutoRemindersDisabled != value {
//			t.CreatorAutoRemindersDisabled = value
//			return true
//		}
//	case t.CounterpartyUserID:
//		if t.CounterpartyAutoRemindersDisabled != value {
//			t.CounterpartyAutoRemindersDisabled = value
//			return true
//		}
//	default:
//		panic("Attempt to set remindersDisabled for a user not related to the transfer")
//	}
//	return false
//}
//
//// Returns true if value have been changed and false if unchanged.
//func (t *Transfer) EnableAutoReminders(userID int64) bool {
//	return t.setAutoRemindersDisabled(userID, false)
//}
//
//// Returns true if value have been changed and false if unchanged.
//func (t *Transfer) DisableAutoReminders(userID int64) bool {
//	return t.setAutoRemindersDisabled(userID, true)
//}

func (t *TransferEntity) Load(ps []datastore.Property) error {
	// Load I and J as usual.
	p2 := make([]datastore.Property, 0, len(ps))
	var creationPlatform string
	var ( // TODO: obsolete props migrated to TransferCounterpartyJson
		creatorReminderID, counterpartyReminderID int64
		creatorTgChatID, counterpartyTgChatID     int64
		creatorTgBotID, counterpartyTgBotID       string
	)
	for _, p := range ps {
		switch p.Name {
		case "AmountInCentsOutstanding": // Ignore legacy
			t.hasObsoleteProps = true
		case "CounterpartyAutoRemindersDisabled": // Ignore legacy
			t.hasObsoleteProps = true
		case "CreatorAutoRemindersDisabled": // Ignore legacy
			t.hasObsoleteProps = true
		case "ReturnTransferIDs": // Ignore legacy
			t.hasObsoleteProps = true
		case "IsDue2Notify": // Ignore legacy
			t.hasObsoleteProps = true
		case "DtDueNext": // Ignore legacy
			t.hasObsoleteProps = true
		case "CounterpartyNotifications": // Ignore legacy
			t.hasObsoleteProps = true
		case "CreatorNotifications": // Ignore legacy
			t.hasObsoleteProps = true
		case "CounterpartyTgReceiptInlineMessageID": // Ignore legacy
			t.hasObsoleteProps = true
		case "CreationPlatform":
			t.hasObsoleteProps = true
			creationPlatform = p.Value.(string)

			//case "FromUserID": // TODO: Ignore legacy, temporary
			//case "FromUserName": // TODO: Ignore legacy, temporary
			//case "FromCounterpartyID": // TODO: Ignore legacy, temporary
			//case "FromCounterpartyName": // TODO: Ignore legacy, temporary
			//case "FromComment": // TODO: Ignore legacy, temporary
			//case "FromNote": // TODO: Ignore legacy, temporary
			//case "ToUserID": // TODO: Ignore legacy, temporary
			//case "ToUserName": // TODO: Ignore legacy, temporary
			//case "ToCounterpartyID": // TODO: Ignore legacy, temporary
			//case "ToCounterpartyName": // TODO:  Ignore legacy, temporary
			//case "ToComment": // TODO: Ignore legacy, temporary
			//case "ToNote": // TODO: Ignore legacy, temporary

		case "CreatorReminderID":
			t.hasObsoleteProps = true
			creatorReminderID = p.Value.(int64)
		case "CounterpartyReminderID":
			t.hasObsoleteProps = true
			counterpartyReminderID = p.Value.(int64)
		case "CreatorTgBotID":
			t.hasObsoleteProps = true
			creatorTgBotID = p.Value.(string)
		case "CounterpartyTgBotID":
			t.hasObsoleteProps = true
			counterpartyTgBotID = p.Value.(string)
		case "CreatorTgChatID":
			t.hasObsoleteProps = true
			creatorTgChatID = p.Value.(int64)
		case "CounterpartyTgChatID":
			t.hasObsoleteProps = true
			counterpartyTgChatID = p.Value.(int64)
		case "Amount":
			t.hasObsoleteProps = true
			if v := p.Value.(float64); v != 0 {
				p.Name = "AmountInCents"
				p.Value = int64(decimal.NewDecimal64p2FromFloat64(v))
				p2 = append(p2, p)
			}
		case "AmountReturned":
			t.hasObsoleteProps = true
			if v := p.Value.(float64); v != 0 {
				p.Name = "AmountInCentsReturned"
				p.Value = int64(decimal.NewDecimal64p2FromFloat64(v))
				p2 = append(p2, p)
			}
		case "AmountOutstanding":
			t.hasObsoleteProps = true
			if v := p.Value.(float64); v != 0 {
				p.Name = "AmountInCentsOutstanding"
				p.Value = int64(decimal.NewDecimal64p2FromFloat64(v))
				p2 = append(p2, p)
			}
		default:
			p2 = append(p2, p)
		}
	}
	if err := datastore.LoadStruct(t, p2); err != nil {
		return err
	}

	if t.CreatedOnPlatform == "" && creationPlatform != "" {
		t.CreatedOnPlatform = creationPlatform
	}

	switch t.DirectionObsoleteProp {
	case "from":
		t.DirectionObsoleteProp = TransferDirectionUser2Counterparty
	case "to":
		t.DirectionObsoleteProp = TransferDirectionCounterparty2User
	}

	if t.GetOutstandingValue(time.Now()) > 0 && !t.IsOutstanding {
		t.IsOutstanding = true
	}

	{ // TODO: Get rid once all transfers migrated - Moves properties to JSON
		migrateToCounterpartyInfo := func(
			counterparty *TransferCounterpartyInfo,
			reminderID, tgChatID int64,
			tgBotID string,
		) {
			if reminderID != 0 {
				t.hasObsoleteProps = true
				counterparty.ReminderID = reminderID
			}
			if tgChatID != 0 {
				t.hasObsoleteProps = true
				counterparty.TgChatID = tgChatID
			}
			if tgBotID != "" {
				t.hasObsoleteProps = true
				counterparty.TgBotID = tgBotID
			}
		}
		migrateToCounterpartyInfo(t.Creator(), creatorReminderID, creatorTgChatID, creatorTgBotID)
		migrateToCounterpartyInfo(t.Counterparty(), counterpartyReminderID, counterpartyTgChatID, counterpartyTgBotID)
	}

	return nil
}

var transferPropertiesToClean = map[string]gaedb.IsOkToRemove{
	// Remove obsolete properties
	"Amount":            gaedb.IsObsolete,
	"AmountTotal":       gaedb.IsObsolete,
	"AmountReturned":    gaedb.IsObsolete,
	"AmountOutstanding": gaedb.IsObsolete,
	//

	// Remove defaults
	"SmsCount":          gaedb.IsZeroInt,
	"SmsCost":           gaedb.IsZeroFloat,
	"SmsCostUSD":        gaedb.IsZeroInt,
	"ReceiptsSentCount": gaedb.IsZeroInt,
	//"CreatorReminderID":         gaedb.IsZeroInt,
	//"CounterpartyReminderID":    gaedb.IsZeroInt,
	//"CreatorTgChatID":           gaedb.IsZeroInt,
	//"CounterpartyTgChatID":      gaedb.IsZeroInt,
	"CreatorTgReceiptByTgMsgID": gaedb.IsZeroInt,
	//"CounterpartyTgBotID":       gaedb.IsEmptyString,
	//"CreatorTgBotID":            gaedb.IsEmptyString,
	"Direction":                gaedb.IsEmptyString,
	"BillID":                   gaedb.IsEmptyString,
	"AmountInCentsOutstanding": gaedb.IsZeroInt,
	"AmountInCentsReturned":    gaedb.IsZeroInt,
	"AcknowledgeStatus":        gaedb.IsEmptyString,
	"AcknowledgeTime":          gaedb.IsZeroTime,
	"DtDueOn":                  gaedb.IsZeroTime,
	"IsOutstanding":            gaedb.IsFalse,
	"IsReturn":                 gaedb.IsFalse,
}

func (t *TransferEntity) Save() (properties []datastore.Property, err error) {
	if t.CreatorUserID == 0 {
		err = errors.New("*TransferEntity.CreatorUserID == 0")
		return
	}

	if t.AmountInCents == 0 { // Should be always presented
		err = errors.New("*TransferEntity.AmountInCents == 0")
		return
	}

	if t.Currency == "" { // Should be always presented
		err = errors.New("*TransferEntity.Currency is empty string")
		return
	}

	if t.AmountInCentsReturned < 0 {
		err = fmt.Errorf("*TransferEntity.AmountInCentsReturned:%v < 0", t.AmountInCentsReturned)
		return
	}

	if err = t.validateTransferInterestAndReturns(); err != nil {
		return
	}

	if t.IsOutstanding {
		switch t.HasInterest() {
		case true:
			// Can we simply check for zero outstanding value?
			// What if there is complex interest rule that allocate interest after grace period?
			if t.GetOutstandingValue(time.Now()) == 0 {
				t.IsOutstanding = false
			}
		case false:
			if t.AmountInCents == t.AmountInCentsReturned {
				t.IsOutstanding = false
			}
		}
	}

	//t.onSaveMigrateUserProps()

	//switch t.Direction() { // TODO: Delete later!
	//case "":
	//	if t.BillID == "" && t.From().UserID == 0 && t.To().UserID == 0 {
	//		err = errors.New("t.Direction is empty string")
	//		return
	//	}
	//case TransferDirectionUser2Counterparty:
	//case TransferDirectionCounterparty2User:
	//default:
	//	err = errors.New("Unknown direction: " + t.DirectionObsoleteProp)
	//	return
	//}

	//if t.AmountInCentsOutstanding < 0 {
	//	err = fmt.Errorf("*TransferEntity.AmountInCentsOutstanding:%v < 0", t.AmountInCentsOutstanding)
	//	return
	//}

	//if t.AmountInCentsReturned > t.AmountInCents {
	//	err = fmt.Errorf("*TransferEntity.AmountInCentsReturned:%v > AmountInCents:%v", t.AmountInCentsReturned, t.AmountInCents)
	//	return
	//}

	//if t.AmountInCentsOutstanding > t.AmountInCents {
	//	err = fmt.Errorf("*TransferEntity.AmountInCentsOutstanding:%v > AmountInCents:%v", t.AmountInCentsOutstanding, t.AmountInCents)
	//	return
	//}
	//
	//if t.AmountInCentsReturned+t.AmountInCentsOutstanding > t.AmountInCents {
	//	err = fmt.Errorf("*TransferEntity.AmountInCentsReturned:%v + AmountInCentsOutstanding:%v > AmountInCents:%v", t.AmountInCentsReturned, t.AmountInCentsOutstanding, t.AmountInCents)
	//	return
	//}

	if t.IsReturn {
		if len(t.ReturnToTransferIDs) == 0 {
			err = errors.New("*TransferEntity: IsReturn == true && len(ReturnToTransferIDs) == 0")
			return
		}
		//if (t.AmountInCentsReturned != 0 || t.AmountInCentsOutstanding != 0) && t.AmountInCents != t.AmountInCentsReturned+t.AmountInCentsOutstanding {
		//	err = fmt.Errorf("*TransferEntity: IsReturn == true && AmountInCents != AmountInCentsReturned + AmountInCentsOutstanding: %v != %v + %v", t.AmountInCents, t.AmountInCentsReturned, t.AmountInCentsOutstanding)
		//	return
		//}
		//} else {
		//	if t.AmountInCents != t.AmountInCentsReturned+t.AmountInCentsOutstanding {
		//		err = fmt.Errorf("*TransferEntity: IsReturn == false && AmountInCents != AmountInCentsReturned + AmountInCentsOutstanding: %v != %v + %v", t.AmountInCents, t.AmountInCentsReturned, t.AmountInCentsOutstanding)
		//		return
		//	}
	}

	if t.CreatorUserID <= 0 { // Should be always presented
		err = fmt.Errorf("*TransferEntity.CreatorUserID:%d <= 0", t.CreatorUserID)
		return
	}

	from := t.From()
	to := t.To()
	if from.UserName == NO_NAME {
		from.UserName = ""
	}
	if to.UserName == NO_NAME {
		to.UserName = ""
	}

	if from.ContactID == 0 && to.ContactID == 0 {
		err = errors.New("from.ContactID == 0 && to.ContactID == 0")
		return
	} else { // Always store 2 values, even if 1 is zero so we can query such records.
		t.BothCounterpartyIDs = []int64{from.ContactID, to.ContactID}
	}

	if from.UserID == 0 && to.UserID == 0 {
		if len(t.BillIDs) == 0 {
			err = errors.New("t.BillIDs is empty && t.From().UserID == 0 && t.To().UserID == 0")
			return
		}
		t.BothUserIDs = []int64{}
	} else { // Always store 2 values, even if 1 is zero so we can query such records.
		t.BothUserIDs = []int64{from.UserID, to.UserID}
	}

	if from.UserID != t.CreatorUserID && from.ContactName == "" && from.UserName == "" { // Should be always presented
		err = errors.New("Either FromCounterpartyName or FromUserName should be presented")
		return
	}
	if to.UserID != t.CreatorUserID && to.ContactName == "" && to.UserName == "" { // Should be always presented
		err = errors.New("Either ToCounterpartyName or ToUserName should be presented")
		return
	}

	if isFixed, s := fixContactName(from.ContactName); isFixed {
		from.ContactName = s
	}

	if isFixed, s := fixContactName(to.ContactName); isFixed {
		to.ContactName = s
	}

	if err = t.onSaveSerializeJson(); err != nil {
		return
	}

	if t.C_From == "" && t.DirectionObsoleteProp == "" {
		err = errors.New("C_From is empty")
		return
	}

	if t.C_To == "" && t.DirectionObsoleteProp == "" {
		err = errors.New("C_To is empty")
		return
	}

	// Serialize from struct to list of properties
	if properties, err = datastore.SaveStruct(t); err != nil {
		return properties, err
	}

	// To optimize storage we filter out default values
	if properties, err = gaedb.CleanProperties(properties, transferPropertiesToClean); err != nil {
		return
	}

	// Obsolete properties also should be removed
	{
		properties2 := make([]datastore.Property, 0, len(properties))
		for _, p := range properties {
			if t.DirectionObsoleteProp == "" && t.C_From != "" && t.C_To != "" &&
				(p.Name == "CreatorCounterpartyID" ||
					p.Name == "CreatorCounterpartyName" ||
					p.Name == "CreatorNote" ||
					p.Name == "CreatorComment" ||
					p.Name == "CounterpartyUserID" ||
					p.Name == "CounterpartyCounterpartyID" ||
					p.Name == "CounterpartyCounterpartyName" ||
					p.Name == "CounterpartyNote" ||
					p.Name == "CounterpartyComment" ||
					p.Name == "DirectionObsoleteProp") {
				continue
			}
			properties2 = append(properties2, p)
		}
		properties = properties2
	}

	// Make general application-wide checks and call hooks if any
	checkHasProperties(TransferKind, properties)

	return
}

func NewTransferEntity(creatorUserID int64, isReturn bool, amount Amount, from *TransferCounterpartyInfo, to *TransferCounterpartyInfo) *TransferEntity {
	if creatorUserID == 0 {
		panic("creatorUserID == 0")
	}
	if from == nil {
		panic("from == nil")
	}
	if to == nil {
		panic("to == nil")
	}
	if amount.Value == 0 {
		panic("amount.Value == 0")
	}
	if amount.Currency == "" {
		panic("amount.Currency is empty")
	}
	transfer := &TransferEntity{
		CreatorUserID: creatorUserID,
		IsReturn:      isReturn,
		//
		from: from,
		to:   to,

		DtCreated: time.Now(),
		//
		//DirectionObsoleteProp: string(direction),
		AmountInCents: amount.Value,
		Currency:      amount.Currency,
	}
	if !isReturn {
		//transfer.AmountInCentsOutstanding = amount.Value
		transfer.IsOutstanding = true
	}
	return transfer
}

func (t *TransferEntity) GetAmount() Amount {
	return Amount{Currency: t.Currency, Value: t.AmountInCents}
}

func (t *TransferEntity) GetReturnedAmount() Amount {
	return Amount{Currency: t.Currency, Value: t.AmountInCentsReturned}
}

func ReverseTransfers(t []Transfer) {
	last := len(t) - 1
	for i := 0; i < len(t)/2; i++ {
		t[i], t[last-i] = t[last-i], t[i]
	}
}
