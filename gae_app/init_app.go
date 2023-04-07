package gaeapp

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/api"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/api/apigaedepended"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/apps/vkapp"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal/gaedal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/reminders"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/support"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/webhooks"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/website"
	//"github.com/strongo/app"
	"net/http"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/maintainance"
	"github.com/julienschmidt/httprouter"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// Init initializes debts tracker server
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

	httpRouter.GET("/test-pointer", testModelPointer)
	httpRouter.GET("/Users/astec/", NotFoundSilent)

	maintainance.RegisterMappers()
}

func NotFoundSilent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusNotFound)
}

func InitCronHandlers(router *httprouter.Router) {
	router.HandlerFunc("GET", "/cron/send-reminders", dtdal.HandleWithContext(reminders.CronSendReminders))
}

func InitTaskQueueHandlers(router *httprouter.Router) {
	router.HandlerFunc("POST", "/taskqueu/send-reminder", dtdal.HandleWithContext(reminders.SendReminderHandler)) // TODO: Remove obsolete!
	router.HandlerFunc("POST", "/task-queue/send-reminder", dtdal.HandleWithContext(reminders.SendReminderHandler))
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

func testModelPointer(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
