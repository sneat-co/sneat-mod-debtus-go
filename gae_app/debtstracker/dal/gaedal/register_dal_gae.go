package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"github.com/strongo/app/db"
	"github.com/strongo/app/gaedb"
	"github.com/strongo/bots-framework/core"
	telegram_bot "github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
)

func RegisterDal() {
	dal.DB = gaedb.NewDatabase()
	telegram_bot.DAL.DB = dal.DB
	//
	dal.Contact = NewContactDalGae()
	dal.Transfer = NewTransferDalGae()
	dal.User = NewUserDalGae()
	dal.Bill = newBillDalGae()
	dal.Split = splitDalGae{}
	dal.TgGroup = newTgGroupDalGae()
	dal.BillSchedule = NewBillScheduleDalGae()
	dal.Receipt = NewReceiptDalGae()
	dal.Reminder = NewReminderDalGae()
	dal.UserBrowser = NewUserBrowserDalGae()
	dal.UserGoogle = NewUserGoogleDalGae()
	dal.PasswordReset = NewPasswordResetDalGae()
	dal.Email = NewEmailDalGae()
	dal.UserGooglePlus = NewUserGooglePlusDalGae()
	dal.UserVk = NewUserVkDalGae()
	dal.UserEmail = NewUserEmailGaeDal()
	dal.UserFacebook = NewUserFacebookDalGae()
	dal.LoginPin = NewLoginPinDalGae()
	dal.LoginCode = NewLoginCodeDalGae()
	dal.Twilio = NewTwilioDalGae()
	dal.Invite = NewInviteDalGae()
	dal.Admin = NewAdminDalGae()
	dal.TgChat = NewTgChatDalGae()
	dal.TgUser = NewTgUserDalGae()
	dal.Group = NewGroupDalGae()
	//dal.GroupMember = NewGroupMemberDalGae()
	dal.UserOneSignal = NewUserOneSignalDalGae()
	dal.UserGaClient = NewUserGaClientDalGae()
	dal.Feedback = NewFeedbackDalGae()
	dal.HttpClient = func(c context.Context) *http.Client {
		return urlfetch.Client(c)
	}
	dal.HandleWithContext = func(handler dal.ContextHandler) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			handler(appengine.NewContext(r), w, r)
		}
	}
	//dal.TaskQueue = TaskQueueDalGae{}
	dal.BotHost = ApiBotHost{}
}

type ApiBotHost struct {
}

func (h ApiBotHost) Context(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func (h ApiBotHost) GetHttpClient(c context.Context) *http.Client {
	return dal.HttpClient(c)
}

func (h ApiBotHost) GetBotCoreStores(platform string, appContext bots.BotAppContext, r *http.Request) bots.BotCoreStores {
	panic("Not implemented")
}

func (h ApiBotHost) DB() db.Database {
	return gaedb.NewDatabase()
}
