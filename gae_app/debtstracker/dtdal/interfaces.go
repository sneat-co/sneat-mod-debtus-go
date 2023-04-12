package dtdal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"fmt"
	tgstore "github.com/bots-go-framework/bots-fw-telegram/store"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/crediterra/money"
	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/app"
	"github.com/strongo/decimal"
	"github.com/strongo/gotwilio"
	"math/rand"
	"net/http"
	"regexp"
	"sync"
	"time"
)

type TransferSource interface {
	PopulateTransfer(t *models.TransferEntity)
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
	TransferID     int
	ReturnedAmount decimal.Decimal64p2
}

type RewardDal interface {
	//GetRewardByID(c context.Context, rewardID int64) (reward models.Reward, err error)
	InsertReward(c context.Context, rewardEntity *models.RewardEntity) (reward models.Reward, err error)
}

type TransferDal interface {
	GetTransfersByID(c context.Context, tx dal.ReadTransaction, transferIDs []int) ([]models.Transfer, error)
	LoadTransfersByUserID(c context.Context, userID int64, offset, limit int) (transfers []models.Transfer, hasMore bool, err error)
	LoadTransfersByContactID(c context.Context, contactID int64, offset, limit int) (transfers []models.Transfer, hasMore bool, err error)
	LoadTransferIDsByContactID(c context.Context, contactID int64, limit int, startCursor string) (transferIDs []int, endCursor string, err error)
	LoadOverdueTransfers(c context.Context, userID int64, limit int) (transfers []models.Transfer, err error)
	LoadOutstandingTransfers(c context.Context, periodEnds time.Time, userID, contactID int64, currency money.Currency, direction models.TransferDirection) (transfers []models.Transfer, err error)
	LoadDueTransfers(c context.Context, userID int64, limit int) (transfers []models.Transfer, err error)
	LoadLatestTransfers(c context.Context, offset, limit int) ([]models.Transfer, error)
	DelayUpdateTransferWithCreatorReceiptTgMessageID(c context.Context, botCode string, transferID int, creatorTgChatID, creatorTgReceiptMessageID int64) error
	DelayUpdateTransfersWithCounterparty(c context.Context, creatorCounterpartyID, counterpartyCounterpartyID int64) error
	DelayUpdateTransfersOnReturn(c context.Context, returnTransferID int, transferReturnUpdates []TransferReturnUpdate) (err error)
}

type ReceiptDal interface {
	UpdateReceipt(c context.Context, receipt models.Receipt) error
	GetReceiptByID(c context.Context, id int) (models.Receipt, error)
	MarkReceiptAsSent(c context.Context, receiptID, transferID int, sentTime time.Time) error
	CreateReceipt(c context.Context, receipt *models.ReceiptEntity) (id int64, err error)
	DelayedMarkReceiptAsSent(c context.Context, receiptID, transferID int64, sentTime time.Time) error
	DelayCreateAndSendReceiptToCounterpartyByTelegram(c context.Context, env strongo.Environment, transferID int, userID int64) error
}

var ErrReminderAlreadyRescheduled = errors.New("reminder already rescheduled")

type ReminderDal interface {
	DelayDiscardReminders(c context.Context, transferIDs []int, returnTransferID int) error
	DelayCreateReminderForTransferUser(c context.Context, transferID int, userID int64) error
	SaveReminder(c context.Context, reminder models.Reminder) (err error)
	GetReminderByID(c context.Context, id int) (models.Reminder, error)
	RescheduleReminder(c context.Context, reminderID int64, remindInDuration time.Duration) (oldReminder, newReminder models.Reminder, err error)
	SetReminderStatus(c context.Context, reminderID int64, returnTransferID int, status string, when time.Time) (reminder models.Reminder, err error)
	DelaySetReminderIsSent(c context.Context, reminderID int64, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error
	SetReminderIsSent(c context.Context, reminderID int64, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) error
	SetReminderIsSentInTransaction(c context.Context, reminder models.Reminder, sentAt time.Time, messageIntID int64, messageStrID, locale, errDetails string) (err error)
	GetActiveReminderIDsByTransferID(c context.Context, transferID int64) ([]int64, error)
	GetSentReminderIDsByTransferID(c context.Context, transferID int64) ([]int64, error)
}

type CreateUserData struct {
	//FbUserID     string
	//GoogleUserID string
	//VkUserID     int64
	FirstName  string
	LastName   string
	ScreenName string
	Nickname   string
}

func CreateUserEntity(createUserData CreateUserData) (user *models.AppUserEntity) {
	return &models.AppUserEntity{
		//FbUserID: createUserData.FbUserID,
		//VkUserID: createUserData.VkUserID,
		//GoogleUniqueUserID: createUserData.GoogleUserID,
		ContactDetails: models.ContactDetails{
			FirstName:  createUserData.FirstName,
			LastName:   createUserData.LastName,
			ScreenName: createUserData.ScreenName,
			Nickname:   createUserData.Nickname,
		},
	}
}

type UserDal interface {
	GetUserByStrID(c context.Context, userID string) (models.AppUser, error)
	GetUserByVkUserID(c context.Context, vkUserID int64) (models.AppUser, error)
	CreateAnonymousUser(c context.Context) (models.AppUser, error)
	CreateUser(c context.Context, userEntity *models.AppUserEntity) (models.AppUser, error)
	DelaySetUserPreferredLocale(c context.Context, delay time.Duration, userID int64, localeCode5 string) error
	DelayUpdateUserHasDueTransfers(c context.Context, userID int64) error
	SetLastCurrency(c context.Context, userID int64, currency money.Currency) error
	DelayUpdateUserWithBill(c context.Context, userID, billID string) error
	DelayUpdateUserWithContact(c context.Context, userID, contactID int64) error
}

type PasswordResetDal interface {
	GetPasswordResetByID(c context.Context, id int64) (models.PasswordReset, error)
	CreatePasswordResetByID(c context.Context, entity *models.PasswordResetEntity) (models.PasswordReset, error)
	SavePasswordResetByID(c context.Context, record models.PasswordReset) (err error)
}

type EmailDal interface {
	InsertEmail(c context.Context, entity *models.EmailEntity) (models.Email, error)
	UpdateEmail(c context.Context, tx dal.ReadwriteTransaction, email models.Email) error
	GetEmailByID(c context.Context, id int64) (models.Email, error)
}

type FeedbackDal interface {
	GetFeedbackByID(c context.Context, feedbackID int) (models.Feedback, error)
}

type ContactDal interface {
	GetLatestContacts(whc botsfw.WebhookContext, limit, totalCount int) (contacts []models.Contact, err error)
	InsertContact(c context.Context, tx dal.ReadwriteTransaction, contactEntity *models.ContactEntity) (contact models.Contact, err error)
	//CreateContact(c context.Context, userID int64, contactDetails models.ContactDetails) (contact models.Contact, user models.AppUser, err error)
	//CreateContactWithinTransaction(c context.Context, user models.AppUser, contactUserID, counterpartyCounterpartyID int64, contactDetails models.ContactDetails, balanced money.Balanced) (contact models.Contact, err error)
	//UpdateContact(c context.Context, contactID int64, values map[string]string) (contactEntity *models.ContactEntity, err error)
	GetContactIDsByTitle(c context.Context, tx dal.ReadTransaction, userID int64, title string, caseSensitive bool) (contactIDs []int64, err error)
	GetContactsWithDebts(c context.Context, tx dal.ReadTransaction, userID int64) (contacts []models.Contact, err error)
}

type BillsHolderGetter func(c context.Context) (billsHolder dal.Record, err error)

type BillDal interface {
	SaveBill(c context.Context, bill models.Bill) (err error)
	UpdateBillsHolder(c context.Context, billID string, getBillsHolder BillsHolderGetter) (err error)
}

type SplitDal interface {
	GetSplitByID(c context.Context, splitID int64) (split models.Split, err error)
	InsertSplit(c context.Context, splitEntity models.SplitEntity) (split models.Split, err error)
}

type TgGroupDal interface {
	GetTgGroupByID(c context.Context, id int64) (tgGroup models.TgGroup, err error)
	SaveTgGroup(c context.Context, tgGroup models.TgGroup) (err error)
}

type BillScheduleDal interface {
	GetBillScheduleByID(c context.Context, id int64) (billSchedule models.BillSchedule, err error)
	InsertBillSchedule(c context.Context, billScheduleEntity *models.BillScheduleEntity) (billSchedule models.BillSchedule, err error)
	UpdateBillSchedule(c context.Context, billSchedule models.BillSchedule) (err error)
}

type GroupDal interface {
	GetGroupByID(c context.Context, tx dal.ReadSession, groupID string) (group models.Group, err error)
	InsertGroup(c context.Context, tx dal.ReadwriteTransaction, groupEntity *models.GroupEntity) (group models.Group, err error)
	SaveGroup(c context.Context, tx dal.ReadwriteTransaction, group models.Group) (err error)
	DelayUpdateGroupWithBill(c context.Context, groupID, billID string) error
}

//type GroupMemberDal interface {
//	GetGroupMemberByID(c context.Context, groupMemberID int64) (groupMember models.GroupMember, err error)
//	CreateGroupMember(c context.Context, groupMemberEntity *models.GroupMemberEntity) (groupMember models.GroupMember, err error)
//}

type UserGoogleDal interface {
	GetUserGoogleByID(c context.Context, googleUserID string) (userGoogle models.UserGoogle, err error)
	DeleteUserGoogle(c context.Context, googleUserID string) (err error)
	SaveUserGoogle(c context.Context, userGoogle models.UserGoogle) (err error)
}

type UserVkDal interface {
	GetUserVkByID(c context.Context, vkUserID int64) (userGoogle models.UserVk, err error)
	SaveUserVk(c context.Context, userVk models.UserVk) (err error)
}

type UserEmailDal interface {
	GetUserEmailByID(c context.Context, email string) (userEmail models.UserEmail, err error)
	SaveUserEmail(c context.Context, userEmail models.UserEmail) (err error)
}

type UserGooglePlusDal interface {
	GetUserGooglePlusByID(c context.Context, id string) (userGooglePlus models.UserGooglePlus, err error)
	SaveUserGooglePlusByID(c context.Context, userGooglePlus models.UserGooglePlus) (err error)
}

type UserFacebookDal interface {
	GetFbUserByFbID(c context.Context, fbAppOrPageID, fbUserOrPageScopeID string) (fbUser models.UserFacebook, err error)
	SaveFbUser(c context.Context, fbUser models.UserFacebook) (err error)
	DeleteFbUser(c context.Context, fbAppOrPageID, fbUserOrPageScopeID string) (err error)
	//CreateFbUserRecord(c context.Context, fbUserID string, appUserID int64) (fbUser models.UserFacebook, err error)
}

type LoginPinDal interface {
	GetLoginPinByID(c context.Context, tx dal.ReadTransaction, loginID int64) (loginPin models.LoginPin, err error)
	SaveLoginPin(c context.Context, tx dal.ReadwriteTransaction, loginPin models.LoginPin) (err error)
	CreateLoginPin(c context.Context, tx dal.ReadwriteTransaction, channel, gaClientID string, createdUserID int64) (int64, error)
}

type LoginCodeDal interface {
	NewLoginCode(c context.Context, userID int64) (int32, error)
	ClaimLoginCode(c context.Context, code int32) (userID int64, err error)
}

type TwilioDal interface {
	GetLastTwilioSmsesForUser(c context.Context, userID int64, to string, limit int) (result []models.TwilioSms, err error)
	SaveTwilioSms(
		c context.Context,
		smsResponse *gotwilio.SmsResponse,
		transfer models.Transfer,
		phoneContact models.PhoneContact,
		userID int64,
		tgChatID int64,
		smsStatusMessageID int,
	) (twiliosSms models.TwilioSms, err error)
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
	GetInvite(c context.Context, inviteCode string) (*models.InviteEntity, error)
	ClaimInvite(c context.Context, userID int64, inviteCode, claimedOn, claimedVia string) (err error)
	ClaimInvite2(c context.Context, inviteCode string, inviteEntity *models.InviteEntity, claimedByUserID int64, claimedOn, claimedVia string) (invite models.Invite, err error)
	CreatePersonalInvite(ec strongo.ExecutionContext, userID int64, inviteBy models.InviteBy, inviteToAddress, createdOnPlatform, createdOnID, related string) (models.Invite, error)
	CreateMassInvite(ec strongo.ExecutionContext, userID int64, inviteCode string, maxClaimsCount int32, createdOnPlatform string) (invite models.Invite, err error)
}

type AdminDal interface {
	DeleteAll(c context.Context, botCode, botChatID string) error
	LatestUsers(c context.Context) (users []models.AppUser, err error)
}

type UserBrowserDal interface {
	SaveUserBrowser(c context.Context, userID int64, userAgent string) (userBrowser models.UserBrowser, err error)
}

type UserOneSignalDal interface {
	SaveUserOneSignal(c context.Context, userID int64, oneSignalUserID string) (userOneSignal models.UserOneSignal, err error)
}

type UserGaClientDal interface {
	SaveGaClient(c context.Context, gaClientId, userAgent, ipAddress string) (gaClient models.GaClient, err error)
}

type TgChatDal interface {
	GetTgChatByID(c context.Context, tgBotID string, tgChatID int64) (tgChat models.TelegramChat, err error)
	DoSomething(c context.Context, // TODO: WTF name?
		userTask *sync.WaitGroup, currency string, tgChatID int64, authInfo auth.AuthInfo, user models.AppUser,
		sendToTelegram func(tgChat tgstore.Chat) error) (err error)
}

type TgUserDal interface {
	FindByUserName(c context.Context, userName string) (tgUsers []tgstore.TgUser, err error)
}

//type TaskQueueDal interface {
//	CallDelayFunc(c context.Context, queueName, subPath, key string, f interface{}, args ...interface{}) error
//}

var (
	DB             dal.Database
	Contact        ContactDal
	User           UserDal
	UserFacebook   UserFacebookDal
	UserGoogle     UserGoogleDal
	UserGooglePlus UserGooglePlusDal
	UserVk         UserVkDal
	PasswordReset  PasswordResetDal
	Email          EmailDal
	UserEmail      UserEmailDal
	UserBrowser    UserBrowserDal
	UserOneSignal  UserOneSignalDal
	UserGaClient   UserGaClientDal
	Feedback       FeedbackDal
	Bill           BillDal
	Split          SplitDal
	BillSchedule   BillScheduleDal
	Receipt        ReceiptDal
	Group          GroupDal
	Reminder       ReminderDal
	TgGroup        TgGroupDal
	Transfer       TransferDal
	Reward         RewardDal
	LoginPin       LoginPinDal
	LoginCode      LoginCodeDal
	Twilio         TwilioDal
	Invite         InviteDal
	Admin          AdminDal
	TgChat         TgChatDal
	TgUser         TgUserDal
	HttpClient     func(c context.Context) *http.Client
	BotHost        botsfw.BotHost
	//TaskQueue		   TaskQueueDal
	HandleWithContext strongo.HandleWithContext
)

func InsertWithRandomStringID(c context.Context, tx dal.ReadwriteTransaction, record dal.Record) error {
	_, _, _ = c, tx, record
	return errors.New("TODO: use dalgo")
}
