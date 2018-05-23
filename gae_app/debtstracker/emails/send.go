package emails

import (
	"bytes"
	"net/http"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/pkg/errors"
	"github.com/strongo/app"
	"github.com/strongo/log"
	"google.golang.org/appengine/urlfetch"
)

func CreateEmailRecordAndQueueForSending(c context.Context, emailEntity *models.EmailEntity) (id int64, err error) {
	var email models.Email

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		emailEntity.Status = "queued"
		if email, err = dal.Email.InsertEmail(c, emailEntity); err != nil {
			err = errors.WithMessage(err, "Failed to insert Email record")
			return err
		}
		if err = DelaySendEmail(c, email.ID); err != nil {
			err = errors.WithMessage(err, "Failed to delay sending")
		}
		return err
	}, dal.CrossGroupTransaction); err != nil {
		return
	}

	return email.ID, err
}

func GetEmailText(c context.Context, translator strongo.SingleLocaleTranslator, templateName string, templateParams interface{}) (string, error) {
	return common.TextTemplates.RenderTemplate(c, translator, templateName, templateParams)
}

func GetEmailHtml(c context.Context, translator strongo.SingleLocaleTranslator, templateName string, templateParams interface{}) (s string, err error) {
	var buffer bytes.Buffer
	err = common.HtmlTemplates.RenderTemplate(c, &buffer, translator, templateName, templateParams)
	return buffer.String(), err
}

func SendEmail(c context.Context, from, to, subject, bodyText, bodyHtml string) (awsSesMessageID string, err error) {
	if bodyText == "" && bodyHtml == "" {
		panic(`bodyText == "" && bodyHtml == ""`)
	}
	svc := ses.New(common.AwsSessionInstance)
	params := &ses.SendEmailInput{
		Destination: &ses.Destination{ // Required
			ToAddresses: []*string{
				aws.String(to), // Required
			},
		},
		Message: &ses.Message{ // Required
			Body: &ses.Body{ // Required
			},
			Subject: &ses.Content{ // Required
				Data:    aws.String(subject), // Required
				Charset: aws.String("utf-8"),
			},
		},
		Source: aws.String(from), // Required
		ReplyToAddresses: []*string{
			aws.String(from), // Required
			// More values...
		},
		//ReturnPath:    aws.String("Address"),
		//ReturnPathArn: aws.String("AmazonResourceName"),
		//SourceArn:     aws.String("AmazonResourceName"),
	}
	if bodyText != "" {
		params.Message.Body.Text = &ses.Content{
			Data:    aws.String(bodyText), // Required
			Charset: aws.String("utf-8"),
		}
	}
	if bodyHtml != "" {
		params.Message.Body.Html = &ses.Content{
			Data:    aws.String(bodyHtml), // Required
			Charset: aws.String("utf-8"),
		}
	}

	http.DefaultClient = urlfetch.Client(c)
	http.DefaultTransport = &urlfetch.Transport{Context: c, AllowInvalidServerCertificate: false}
	log.Debugf(c, "Sending email through AWS SES: %v", params)

	resp, err := svc.SendEmail(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the ByCode and
		// Message from an error.
		originalErr := err
		errMessage := err.Error()
		if to != strings.ToLower(to) && strings.Index(errMessage, "Email address is not verified") > 0 && strings.Index(errMessage, to) > 0 {
			params.Destination.ToAddresses[0] = aws.String(strings.ToLower(to))
			resp, err = svc.SendEmail(params)
			if err != nil {
				log.Errorf(c, "Failed to send ToLower(email): %v", err)
				return "", originalErr
			}
		} else {
			log.Errorf(c, "Failed to send email using AWS SES: %v", err)
			return "", errors.Wrap(err, "Failed to send email")
		}
	}

	// Pretty-print the response data.
	log.Debugf(c, "AWS SES output: %v", resp)
	return *resp.MessageId, err
}
