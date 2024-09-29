package dtdal

import (
	"context"
	"errors"
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-core-modules/auth/models4auth"
	"github.com/sneat-co/sneat-core-modules/contactus/dto4contactus"
	"github.com/sneat-co/sneat-core-modules/userus/dbo4userus"
	models4debtus2 "github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	models4splitus2 "github.com/sneat-co/sneat-mod-debtus-go/splitus/models4splitus"
	"github.com/strongo/decimal"
	"github.com/strongo/gotwilio"
	"github.com/strongo/strongoapp"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

type TransferSource interface {
	PopulateTransfer(t *models4debtus2.TransferData)
}

const (
	AckAccept  = "accept"
	AckDecline = "decline"
)

//var (
//	CrossGroupTransaction  = dal.CrossGroupTransaction
//	SingleGroupTransaction = db.SingleGroupTransaction
//)

type TransferReturnUpdate struct {
	TransferID     string
	ReturnedAmount decimal.Decimal64p2
}

type RewardDal interface {
	//GetRewardByID(ctx context.Context, rewardID int64) (reward models.Reward, err error)
	InsertReward(ctx context.Context, tx dal.ReadwriteTransaction, rewardEntity *models4debtus2.RewardDbo) (reward models4debtus2.Reward, err error)
}

type TransferDal interface {
	GetTransfersByID(ctx context.Context, tx dal.ReadSession, transferIDs []string) ([]models4debtus2.TransferEntry, error)
	LoadTransfersByUserID(ctx context.Context, userID string, offset, limit int) (transfers []models4debtus2.TransferEntry, hasMore bool, err error)
	LoadTransfersByContactID(ctx context.Context, contactID string, offset, limit int) (transfers []models4debtus2.TransferEntry, hasMore bool, err error)
	LoadTransferIDsByContactID(ctx context.Context, contactID string, limit int, startCursor string) (transferIDs []string, endCursor string, err error)
	LoadOverdueTransfers(ctx context.Context, tx dal.ReadSession, userID string, limit int) (transfers []models4debtus2.TransferEntry, err error)
	LoadOutstandingTransfers(ctx context.Context, tx dal.ReadSession, periodEnds time.Time, userID, contactID string, currency money.CurrencyCode, direction models4debtus2.TransferDirection) (transfers []models4debtus2.TransferEntry, err error)
	LoadDueTransfers(ctx context.Context, tx dal.ReadSession, userID string, limit int) (transfers []models4debtus2.TransferEntry, err error)
	LoadLatestTransfers(ctx context.Context, offset, limit int) ([]models4debtus2.TransferEntry, error)
	DelayUpdateTransferWithCreatorReceiptTgMessageID(ctx context.Context, botCode string, transferID string, creatorTgChatID, creatorTgReceiptMessageID int64) error
	DelayUpdateTransfersWithCounterparty(ctx context.Context, creatorCounterpartyID, counterpartyCounterpartyID string) error
	DelayUpdateTransfersOnReturn(ctx context.Context, returnTransferID string, transferReturnUpdates []TransferReturnUpdate) (err error)
}

type ReceiptDal interface {
	UpdateReceipt(ctx context.Context, tx dal.ReadwriteTransaction, receipt models4debtus2.ReceiptEntry) error
	GetReceiptByID(ctx context.Context, tx dal.ReadSession, id string) (models4debtus2.ReceiptEntry, error)
	MarkReceiptAsSent(ctx context.Context, receiptID, transferID string, sentTime time.Time) error
	CreateReceipt(ctx context.Context, data *models4debtus2.ReceiptDbo) (receipt models4debtus2.ReceiptEntry, err error)
	DelayedMarkReceiptAsSent(ctx context.Context, receiptID, transferID string, sentTime time.Time) error
	DelayCreateAndSendReceiptToCounterpartyByTelegram(ctx context.Context, env string, transferID string, userID string) error
}

var ErrReminderAlreadyRescheduled = errors.New("reminder already rescheduled")

type ReminderDal interface {
	DelayDiscardReminders(ctx context.Context, transferIDs []string, returnTransferID string) error
	DelayCreateReminderForTransferUser(ctx context.Context, transferID string, userID string) error
	SaveReminder(ctx context.Context, tx dal.ReadwriteTransaction, reminder models4debtus2.Reminder) (err error)
	GetReminderByID(ctx context.Context, tx dal.ReadSession, id string) (models4debtus2.Reminder, error)
	RescheduleReminder(ctx context.Context, reminderID string, remindInDuration time.Duration) (oldReminder, newReminder models4debtus2.Reminder, err error)
	SetReminderStatus(ctx context.Context, reminderID string, returnTransferID string, status string, when time.Time) (reminder models4debtus2.Reminder, err error)
	DelaySetReminderIsSent(ctx context.Context, reminderID string, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error
	SetReminderIsSent(ctx context.Context, reminderID string, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error
	SetReminderIsSentInTransaction(ctx context.Context, tx dal.ReadwriteTransaction, reminder models4debtus2.Reminder, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) (err error)
	GetActiveReminderIDsByTransferID(ctx context.Context, tx dal.ReadSession, transferID int) ([]int, error)
	GetSentReminderIDsByTransferID(ctx context.Context, tx dal.ReadSession, transferID int) ([]int, error)
}

type FeedbackDal interface {
	GetFeedbackByID(ctx context.Context, tx dal.ReadSession, feedbackID int64) (feedback models4debtus2.Feedback, err error)
}

type ContactDal interface {
	GetLatestContacts(whc botsfw.WebhookContext, tx dal.ReadSession, spaceID string, limit, totalCount int) (contacts []models4debtus2.DebtusSpaceContactEntry, err error)
	InsertContact(ctx context.Context, tx dal.ReadwriteTransaction, contactEntity *models4debtus2.DebtusSpaceContactDbo) (contact models4debtus2.DebtusSpaceContactEntry, err error)
	GetContactIDsByTitle(ctx context.Context, tx dal.ReadSession, spaceID, userID string, title string, caseSensitive bool) (contactIDs []string, err error)
	GetContactsWithDebts(ctx context.Context, tx dal.ReadSession, spaceID, userID string) (contacts []models4debtus2.DebtusSpaceContactEntry, err error)
}

type BillsHolderGetter func(ctx context.Context) (billsHolder dal.Record, err error)

type SplitDal interface {
	GetSplitByID(ctx context.Context, splitID int) (split models4splitus2.Split, err error)
	InsertSplit(ctx context.Context, splitEntity models4splitus2.SplitEntity) (split models4splitus2.Split, err error)
}

type TgGroupDal interface {
	GetTgGroupByID(ctx context.Context, tx dal.ReadSession, id int64) (tgGroup models4auth.TgGroup, err error)
	SaveTgGroup(ctx context.Context, tx dal.ReadwriteTransaction, tgGroup models4auth.TgGroup) (err error)
}

type BillScheduleDal interface {
	GetBillScheduleByID(ctx context.Context, id int64) (billSchedule models4splitus2.BillSchedule, err error)
	InsertBillSchedule(ctx context.Context, billScheduleEntity *models4splitus2.BillScheduleEntity) (billSchedule models4splitus2.BillSchedule, err error)
	UpdateBillSchedule(ctx context.Context, billSchedule models4splitus2.BillSchedule) (err error)
}

//type GroupMemberDal interface {
//	GetGroupMemberByID(ctx context.Context, groupMemberID int64) (groupMember models.GroupMember, err error)
//	CreateGroupMember(ctx context.Context, groupMemberEntity *models.GroupMemberData) (groupMember models.GroupMember, err error)
//}

type TwilioDal interface {
	GetLastTwilioSmsesForUser(ctx context.Context, tx dal.ReadSession, userID string, to string, limit int) (result []models4debtus2.TwilioSms, err error)
	SaveTwilioSms(
		ctx context.Context,
		smsResponse *gotwilio.SmsResponse,
		transfer models4debtus2.TransferEntry,
		phoneContact dto4contactus.PhoneContact,
		userID string,
		tgChatID int64,
		smsStatusMessageID int,
	) (twiliosSms models4debtus2.TwilioSms, err error)
}

const LetterBytes = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed 1, I and 0, O as can be messed with l/1 and 0.
var InviteCodeRegex = regexp.MustCompile(fmt.Sprintf("[%v]+", LetterBytes))

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomCode(n uint8) string {
	b := make([]byte, n)
	lettersCount := len(LetterBytes)
	for i := range b {
		b[i] = LetterBytes[random.Intn(lettersCount)]
	}
	return string(b)
}

type InviteDal interface {
	GetInvite(ctx context.Context, tx dal.ReadSession, inviteCode string) (models4debtus2.Invite, error)
	ClaimInvite(ctx context.Context, userID string, inviteCode, claimedOn, claimedVia string) (err error)
	ClaimInvite2(ctx context.Context, inviteCode string, invite models4debtus2.Invite, claimedByUserID string, claimedOn, claimedVia string) (err error)
	CreatePersonalInvite(ec strongoapp.ExecutionContext, userID string, inviteBy models4debtus2.InviteBy, inviteToAddress, createdOnPlatform, createdOnID, related string) (models4debtus2.Invite, error)
	CreateMassInvite(ec strongoapp.ExecutionContext, userID string, inviteCode string, maxClaimsCount int32, createdOnPlatform string) (invite models4debtus2.Invite, err error)
}

type AdminDal interface {
	DeleteAll(ctx context.Context, botCode, botChatID string) error
	LatestUsers(ctx context.Context) (users []dbo4userus.UserEntry, err error)
}

//type TaskQueueDal interface {
//	CallDelayFunc(ctx context.Context, queueName, subPath, key string, f interface{}, args ...interface{}) error
//}

var (
	DB      dal.DB
	Contact ContactDal

	Feedback FeedbackDal
	//Bill      BillDal
	Receipt  ReceiptDal
	Reminder ReminderDal
	TgGroup  TgGroupDal
	Transfer TransferDal
	Twilio   TwilioDal
	Invite   InviteDal
	Admin    AdminDal

	HttpClient  func(ctx context.Context) *http.Client
	BotHost     botsfw.BotHost
	HttpAppHost strongoapp.HttpAppHost

	//Split        SplitDal
	//BillSchedule BillScheduleDal
	//Reward RewardDal
	//TaskQueue		   TaskQueueDal
	//UserVk         UserVkDal

)

func InsertWithRandomStringID(_ context.Context, tx dal.ReadwriteTransaction, record dal.Record) error {
	_, _ = tx, record
	return errors.New("TODO: use dalgo")
}
