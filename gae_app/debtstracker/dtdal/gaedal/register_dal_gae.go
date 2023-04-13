package gaedal

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	strongo "github.com/strongo/app"
	"github.com/strongo/db"
	"net/http"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"context"
	telegramBot "github.com/bots-go-framework/bots-fw-telegram"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine"
	"google.golang.org/appengine/v2/urlfetch"
)

func RegisterDal() {
	dtdal.DB = gaedb.NewDatabase()
	telegramBot.DAL.DB = dtdal.DB
	//
	dtdal.Contact = NewContactDalGae()
	dtdal.Transfer = NewTransferDalGae()
	dtdal.Reward = NewRewardDalGae()
	dtdal.User = NewUserDalGae()
	dtdal.Bill = newBillDalGae()
	dtdal.Split = splitDalGae{}
	dtdal.TgGroup = newTgGroupDalGae()
	dtdal.BillSchedule = NewBillScheduleDalGae()
	dtdal.Receipt = NewReceiptDalGae()
	dtdal.Reminder = NewReminderDalGae()
	dtdal.UserBrowser = NewUserBrowserDalGae()
	dtdal.UserGoogle = NewUserGoogleDalGae()
	dtdal.PasswordReset = NewPasswordResetDalGae()
	dtdal.Email = NewEmailDalGae()
	dtdal.UserGooglePlus = NewUserGooglePlusDalGae()
	dtdal.UserVk = NewUserVkDalGae()
	dtdal.UserEmail = NewUserEmailGaeDal()
	dtdal.UserFacebook = NewUserFacebookDalGae()
	dtdal.LoginPin = NewLoginPinDalGae()
	dtdal.LoginCode = NewLoginCodeDalGae()
	dtdal.Twilio = NewTwilioDalGae()
	dtdal.Invite = NewInviteDalGae()
	dtdal.Admin = NewAdminDalGae()
	dtdal.TgChat = NewTgChatDalGae()
	dtdal.TgUser = NewTgUserDalGae()
	dtdal.Group = NewGroupDalGae()
	//dtdal.GroupMember = NewGroupMemberDalGae()
	dtdal.UserOneSignal = NewUserOneSignalDalGae()
	dtdal.UserGaClient = NewUserGaClientDalGae()
	dtdal.Feedback = NewFeedbackDalGae()
	dtdal.HttpClient = func(c context.Context) *http.Client {
		return urlfetch.Client(c)
	}
	dtdal.HandleWithContext = func(handler strongo.ContextHandler) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			handler(appengine.NewContext(r), w, r)
		}
	}
	//dtdal.TaskQueue = TaskQueueDalGae{}
	dtdal.BotHost = ApiBotHost{}
}

type ApiBotHost struct {
}

func (h ApiBotHost) Context(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func (h ApiBotHost) GetHTTPClient(c context.Context) *http.Client {
	return dtdal.HttpClient(c)
}

func (h ApiBotHost) GetBotCoreStores(platform string, appContext botsfw.BotAppContext, r *http.Request) botsfw.BotCoreStores {
	panic("Not implemented")
}

func (h ApiBotHost) DB() db.Database {
	return gaedb.NewDatabase()
}
