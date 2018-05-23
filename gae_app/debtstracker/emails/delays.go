package emails

import (
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/pkg/errors"
	"github.com/strongo/app/gae"
	"github.com/strongo/db"
	"github.com/strongo/log"
	"google.golang.org/appengine/delay"
)

const SEND_EMAIL_TASK = "send-email"

func DelaySendEmail(c context.Context, id int64) error {
	return gae.CallDelayFunc(c, common.QUEUE_EMAILS, SEND_EMAIL_TASK, delayEmail, id)
}

var delayEmail = delay.Func(SEND_EMAIL_TASK, delayedSendEmail)

var ErrEmailIsInWrongStatus = errors.New("email is already sending or sent")

func delayedSendEmail(c context.Context, id int64) (err error) {
	log.Debugf(c, "delayedSendEmail(%v)", id)

	var email models.Email

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if email, err = dal.Email.GetEmailByID(c, id); err != nil {
			return err
		}
		if email.Status != "queued" {
			return errors.WithMessage(ErrEmailIsInWrongStatus, "Expected 'queued' got email.Status="+email.Status)
		}
		email.Status = "sending"
		return dal.Email.UpdateEmail(c, email)
	}, dal.SingleGroupTransaction); err != nil {
		err = errors.WithMessage(err, "Failed to update email status to 'queued'")
		if db.IsNotFound(err) {
			log.Warningf(c, err.Error())
			return nil // Do not retry
		} else if errors.Cause(err) == ErrEmailIsInWrongStatus {
			log.Warningf(c, err.Error())
			return nil // Do not retry
		}
		log.Errorf(c, err.Error())
		return err // Retry
	}

	var awsSesMessageID string
	if awsSesMessageID, err = SendEmail(c, email.From, email.To, email.Subject, email.BodyText, email.BodyHtml); err != nil {
		log.Errorf(c, "Failed to send email: %v", err)

		if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
			if email, err = dal.Email.GetEmailByID(c, id); err != nil {
				return err
			}
			if email.Status != "sending" {
				return errors.WithMessage(ErrEmailIsInWrongStatus, "Expected 'sending' got email.Status="+email.Status)
			}
			email.Status = "error"
			email.Error = err.Error()
			return dal.Email.UpdateEmail(c, email)
		}, dal.SingleGroupTransaction); err != nil {
			log.Errorf(c, err.Error())
		}
		return nil // Do not retry
	}

	log.Infof(c, "Sent email, message ID: %v", awsSesMessageID)

	if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
		if email, err = dal.Email.GetEmailByID(c, id); err != nil {
			return err
		}
		if email.Status != "sending" {
			return errors.WithMessage(ErrEmailIsInWrongStatus, "Expected 'sending' got email.Status="+email.Status)
		}
		email.Status = "sent"
		email.DtSent = time.Now()
		email.AwsSesMessageID = awsSesMessageID
		return dal.Email.UpdateEmail(c, email)
	}, dal.SingleGroupTransaction); err != nil {
		log.Errorf(c, err.Error())
		err = nil // Do not retry!
	}
	return nil // Do not retry!
}
