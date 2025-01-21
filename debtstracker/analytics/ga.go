package analytics

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/common4debtus"
	"github.com/strongo/gamp"
	"github.com/strongo/logus"
	"net/http"
)

const (
	BASE_HOST = ".debtusbot.io"
)

const (
	EventCategoryReminders  = "reminders"
	EventActionReminderSent = "reminder-sent"
)

const (
	EventCategoryTransfers    = "api4transfers"
	EventActionDebtDueDateSet = "debt-due-date-set"
)

func SendSingleMessage(ctx context.Context, m gamp.Message) (err error) {
	if ctx == nil {
		return errors.New("parameter 'ctx context.Context' is nil")
	}
	gaMeasurement := gamp.NewBufferedClient("", dtdal.HttpClient(ctx), nil)
	if err = gaMeasurement.Queue(m); err != nil {
		return err
	}
	if err = gaMeasurement.Flush(); err != nil {
		return err
	}
	var buffer bytes.Buffer
	_, _ = m.Write(&buffer)
	logus.Debugf(ctx, "Sent single message to GA: "+buffer.String())
	return nil
}

func getGaCommon(r *http.Request, userID string, userLanguage, platform string) gamp.Common {
	var userAgent string
	if r != nil {
		userAgent = r.UserAgent()
	} else {
		userAgent = "appengine"
	}

	return gamp.Common{
		TrackingID:    common4debtus.GA_TRACKING_ID,
		UserID:        userID,
		UserLanguage:  userLanguage,
		UserAgent:     userAgent,
		DataSource:    "backend",
		ApplicationID: "io.debtusbot.gae",
	}
}

func ReminderSent(ctx context.Context, userID string, userLanguage, platform string) {
	gaCommon := getGaCommon(nil, userID, userLanguage, platform)
	if err := SendSingleMessage(ctx, gamp.NewEvent(EventCategoryReminders, EventActionReminderSent, gaCommon)); err != nil {
		logus.Errorf(ctx, fmt.Errorf("failed to send even to GA: %w", err).Error())
	}
}

func ReceiptSentFromApi(ctx context.Context, r *http.Request, userID string, userLanguage, platform, channel string) {
	gaCommon := getGaCommon(r, userID, userLanguage, platform)
	_ = SendSingleMessage(ctx, gamp.NewEventWithLabel(
		"receipts",
		"receipt-sent",
		channel,
		gaCommon,
	))
}
