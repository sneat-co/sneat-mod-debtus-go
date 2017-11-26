package analytics

import (
	"bytes"
	"net/http"
	"strconv"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
	"github.com/strongo/measurement-protocol"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

const (
	BASE_HOST = ".debtstracker.io"
)

const (
	EventCategory_Reminders  = "reminders"
	EventAction_ReminderSent = "reminder-sent"
)

const (
	EventCategory_Transfers    = "transfers"
	EventAction_DebtDueDateSet = "debt-due-date-set"
)

func SendSingleMessage(c context.Context, m measurement.Message) (err error) {
	if c == nil {
		return errors.New("Parameter 'c context.Context' is nil")
	}
	gaMeasurement := measurement.NewBufferedSender([]string{common.GA_TRACKING_ID}, true, urlfetch.Client(c))
	if err = gaMeasurement.Queue(m); err != nil {
		return err
	}
	if err = gaMeasurement.Flush(); err != nil {
		return err
	}
	var buffer bytes.Buffer
	m.Write(&buffer, common.GA_TRACKING_ID)
	log.Debugf(c, "Sent single messasge to GA: "+buffer.String())
	return nil
}

func getGaCommon(r *http.Request, userID int64, userLanguage, platform string) measurement.Common {
	var userAgent string
	if r != nil {
		userAgent = r.UserAgent()
	} else {
		userAgent = "appengine"
	}

	return measurement.Common{
		UserID:        strconv.FormatInt(userID, 10),
		UserLanguage:  userLanguage,
		UserAgent:     userAgent,
		DataSource:    "backend",
		ApplicationID: "io.debtstracker.gae",
	}
}

func ReminderSent(c context.Context, userID int64, userLanguage, platform string) {
	gaCommon := getGaCommon(nil, userID, userLanguage, platform)
	if err := SendSingleMessage(c, measurement.NewEvent(EventCategory_Reminders, EventAction_ReminderSent, gaCommon)); err != nil {
		log.Errorf(c, errors.Wrap(err, "Failed to send even to GA").Error())
	}
}

func ReceiptSentFromBot(whc bots.WebhookContext, channel string) {
	whc.GaMeasurement().Queue(whc.GaEventWithLabel("receipts", "receipt-sent", channel))
}

func ReceiptSentFromApi(c context.Context, r *http.Request, userID int64, userLanguage, platform, channel string) {
	gaCommon := getGaCommon(r, userID, userLanguage, platform)
	SendSingleMessage(c, measurement.NewEventWithLabel(
		"receipts",
		"receipt-sent",
		channel,
		gaCommon,
	))
}
