package reminders

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"time"
)

func sendReminderByEmail(c context.Context, reminder models.Reminder, emailTo string, transfer models.Transfer, user models.AppUserEntity) error {
	log.Debugf(c, "sendReminderByEmail(reminder.ID=%v, emailTo=%v)", reminder.ID, emailTo)
	// TODO: Do we really need to pass "w http.ResponseWriter" here?
	var text bytes.Buffer
	var subj bytes.Buffer

	subj.WriteString("Due payment notification")
	text.WriteString(fmt.Sprintf("Hi %v, you have a due payment to %v: %v%v.", transfer.Counterparty().ContactName, user.Username, transfer.AmountInCents, transfer.Currency))

	svc := ses.New(common.AwsSessionInstance)
	params := &ses.SendEmailInput{
		Destination: &ses.Destination{ // Required
			ToAddresses: []*string{
				aws.String(emailTo), // Required
			},
		},
		Message: &ses.Message{ // Required
			Body: &ses.Body{ // Required
				//Html: &ses.Content{
				//	Data:    aws.String(html.String()), // Required
				//	Charset: aws.String("utf-8"),
				//},
				Text: &ses.Content{
					Data:    aws.String(text.String()), // Required
					Charset: aws.String("utf-8"),
				},
			},
			Subject: &ses.Content{ // Required
				Data:    aws.String(subj.String()), // Required
				Charset: aws.String("utf-8"),
			},
		},
		Source: aws.String(common.FROM_REMINDER), // Required
		ReplyToAddresses: []*string{
			aws.String(common.FROM_REMINDER), // Required
			// More values...
		},
		//ReturnPath:    aws.String("Address"),
		//ReturnPathArn: aws.String("AmazonResourceName"),
		//SourceArn:     aws.String("AmazonResourceName"),
	}

	http.DefaultClient = urlfetch.Client(c)
	http.DefaultTransport = &urlfetch.Transport{Context: c, AllowInvalidServerCertificate: false}
	resp, err := svc.SendEmail(params)

	sentAt := time.Now()

	var (
		emailMessageID string
		errDetails     string
	)
	if resp.MessageId != nil {
		emailMessageID = *resp.MessageId
	}

	if err != nil {
		errDetails = err.Error()
	}

	if err = dal.Reminder.SetReminderIsSent(c, reminder.ID, sentAt, 0, emailMessageID, strongo.LOCALE_EN_US, errDetails); err != nil {
		dal.Reminder.DelaySetReminderIsSent(c, reminder.ID, sentAt, 0, emailMessageID, strongo.LOCALE_EN_US, errDetails)
	}

	if err != nil {
		// Print the error, cast err to awserr.Error to get the ByCode and
		// Message from an error.
		return errors.Wrap(err, "Failed to send email using AWS SES")
	}

	// Pretty-print the response data.
	log.Debugf(c, "AWS SES output (for Reminder=%v): %v", reminder.ID, resp)
	return nil
}
