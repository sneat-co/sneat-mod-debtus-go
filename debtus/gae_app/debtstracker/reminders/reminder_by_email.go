package reminders

import (
	"context"
	"fmt"
	"github.com/sneat-co/sneat-core-modules/common4all"
	"github.com/sneat-co/sneat-core-modules/emailing"
	"github.com/sneat-co/sneat-core-modules/userus/dbo4userus"
	"github.com/sneat-co/sneat-go-core/emails"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/gae_app/debtstracker/dtdal"
	models4debtus2 "github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/strongo/i18n"
	"github.com/strongo/logus"
	"time"
)

func sendReminderByEmail(ctx context.Context, reminder models4debtus2.Reminder, emailTo string, transfer models4debtus2.TransferEntry, user dbo4userus.UserEntry) (err error) {
	logus.Debugf(ctx, "sendReminderByEmail(reminder.ContactID=%v, emailTo=%v)", reminder.ID, emailTo)

	emailMessage := emails.Email{
		From: common4all.FROM_REMINDER,
		To: []string{
			emailTo, // Required
		},
		Subject: "Due payment notification",
		Text:    fmt.Sprintf("Hi %v, you have a due payment to %v: %v%v.", transfer.Data.Counterparty().ContactName, user.Data.Names.UserName, transfer.Data.AmountInCents, transfer.Data.Currency),
	}

	var emailClient emails.Client

	if emailClient, err = emailing.GetEmailClient(ctx); err != nil {
		return
	}

	var sent emails.Sent
	sent, err = emailClient.Send(emailMessage)

	sentAt := time.Now()

	var errDetails string
	if err != nil {
		errDetails = err.Error()
	}
	var emailMessageID string
	if sent != nil {
		emailMessageID = sent.MessageID()
	}

	if err = dtdal.Reminder.SetReminderIsSent(ctx, reminder.ID, sentAt, 0, emailMessageID, i18n.LocaleCodeEnUS, errDetails); err != nil {
		if err = dtdal.Reminder.DelaySetReminderIsSent(ctx, reminder.ID, sentAt, 0, emailMessageID, i18n.LocaleCodeEnUS, errDetails); err != nil {
			return fmt.Errorf("failed to delay setting reminder as sent: %w", err)
		}
	}

	if err != nil {
		// Print the error, cast err to awserr.Error to get the ByCode and
		// Message from an error.
		return fmt.Errorf("failed to send email using AWS SES: %w", err)
	}

	// Pretty-print the response data.
	logus.Debugf(ctx, "AWS SES output (for Reminder=%v): %v", reminder.ID, sent)
	return nil
}
