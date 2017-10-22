package gae_app

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/api"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/api/apigaedepended"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/apps/vkapp"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal/gaedal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/reminders"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/support"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/webhooks"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/website"
	//"github.com/strongo/app"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/maintainance"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/app/log"
	"github.com/strongo/bots-framework/core"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
)

func Init(botHost bots.BotHost) {
	if botHost == nil {
		panic("botHost parameter is required")
	}
	gaedal.RegisterDal()
	apigaedepended.InitApiGaeDepended()

	httpRouter := httprouter.New()
	http.Handle("/", httpRouter)

	api.InitApi(httpRouter)
	website.InitWebsite(httpRouter)
	webhooks.InitWebhooks(httpRouter)
	vkapp.InitVkIFrameApp(httpRouter)
	support.InitSupportHandlers(httpRouter)

	InitCronHandlers(httpRouter)
	InitTaskQueueHandlers(httpRouter)

	InitBots(httpRouter, botHost, common.TheAppContext)

	httpRouter.GET("/test-pointer", TestModelPointer)
	httpRouter.GET( "/Users/astec/", NotFoundSilent)

	maintainance.RegisterMappers()
}

func NotFoundSilent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusNotFound)
}

func InitCronHandlers(router *httprouter.Router) {
	router.HandlerFunc("GET", "/cron/send-reminders", dal.HandleWithContext(reminders.CronSendReminders))
}

func InitTaskQueueHandlers(router *httprouter.Router) {
	router.HandlerFunc("POST", "/taskqueu/send-reminder", dal.HandleWithContext(reminders.SendReminderHandler)) // TODO: Remove obsolete!
	router.HandlerFunc("POST", "/task-queue/send-reminder", dal.HandleWithContext(reminders.SendReminderHandler))
}

type TestTransferCounterparty struct {
	UserID   int64  `datastore:",noindex"`
	UserName string `datastore:",noindex"`
	Comment  string `datastore:",noindex"`
}

type TestTransfer struct {
	From TestTransferCounterparty
	To   TestTransferCounterparty
}

func TestModelPointer(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)
	testTransfer := TestTransfer{
		From: TestTransferCounterparty{UserID: 1, UserName: "First"},
		To:   TestTransferCounterparty{UserID: 2, UserName: "Second"},
	}
	key := datastore.NewKey(c, "TestTransfer", "", 1, nil)
	if _, err := datastore.Put(c, key, &testTransfer); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	var testTransfer2 TestTransfer
	datastore.Get(c, key, &testTransfer2)
	log.Debugf(c, "testTransfer2: %v", testTransfer2)
	log.Debugf(c, "testTransfer2.From: %v", testTransfer2.From)
	log.Debugf(c, "testTransfer2.To: %v", testTransfer2.To)
	testTransfer2.From.Comment = "Comment #1"
	if _, err := datastore.Put(c, key, &testTransfer); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	var testTransfer3 TestTransfer
	datastore.Get(c, key, &testTransfer3)
	log.Debugf(c, "testTransfer2: %v", testTransfer3)
	log.Debugf(c, "testTransfer2.From: %v", testTransfer3.From)
	log.Debugf(c, "testTransfer2.To: %v", testTransfer3.To)
}
