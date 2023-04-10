package analytics

import (
	"bytes"
	"net/http"
	"strconv"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"context"
	"errors"
	"github.com/strongo/gamp"
	"github.com/strongo/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	BASE_HOST = ".debtstracker.io"
)

const (
	EventCategoryReminders  = "reminders"
	EventActionReminderSent = "reminder-sent"
)

const (
	EventCategoryTransfers    = "transfers"
	EventActionDebtDueDateSet = "debt-due-date-set"
)

func SendSingleMessage(c context.Context, m gamp.Message) (err error) {
	if c == nil {
		return errors.New("Parameter 'c context.Context' is nil")
	}
	gaMeasurement := gamp.NewBufferedClient("", urlfetch.Client(c), nil)
	if err = gaMeasurement.Queue(m); err != nil {
		return err
	}
	if err = gaMeasurement.Flush(); err != nil {
		return err
	}
	var buffer bytes.Buffer
	m.Write(&buffer)
	log.Debugf(c, "Sent single message to GA: "+buffer.String())
	return nil
}

func getGaCommon(r *http.Request, userID int64, userLanguage, platform string) gamp.Common {
	var userAgent string
	if r != nil {
		userAgent = r.UserAgent()
	} else {
		userAgent = "appengine"
	}

	return gamp.Common{
		TrackingID:    common.GA_TRACKING_ID,
		UserID:        strconv.FormatInt(userID, 10),
		UserLanguage:  userLanguage,
		UserAgent:     userAgent,
		DataSource:    "backend",
		ApplicationID: "io.debtstracker.gae",
	}
}

func ReminderSent(c context.Context, userID int64, userLanguage, platform string) {
	gaCommon := getGaCommon(nil, userID, userLanguage, platform)
	if err := SendSingleMessage(c, gamp.NewEvent(EventCategoryReminders, EventActionReminderSent, gaCommon)); err != nil {
		log.Errorf(c, errors.Wrap(err, "Failed to send even to GA").Error())
	}
}

func ReceiptSentFromBot(whc botsfw.WebhookContext, channel string) error {
	ga := whc.GA()
	return ga.Queue(ga.GaEventWithLabel("receipts", "receipt-sent", channel))
}

func ReceiptSentFromApi(c context.Context, r *http.Request, userID int64, userLanguage, platform, channel string) {
	gaCommon := getGaCommon(r, userID, userLanguage, platform)
	SendSingleMessage(c, gamp.NewEventWithLabel(
		"receipts",
		"receipt-sent",
		channel,
		gaCommon,
	))
}
