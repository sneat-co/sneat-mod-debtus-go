package emails

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"errors"
	"github.com/strongo/app/gae"
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

	var db dal.Database
	if db, err = facade.GetDatabase(c); err != nil {
		return err
	}

	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		if email, err = dtdal.Email.GetEmailByID(c, id); err != nil {
			return err
		}
		if email.Status != "queued" {
			return fmt.Errorf("%w: expected 'queued' got email.Status=%s", ErrEmailIsInWrongStatus, email.Status)
		}
		email.Status = "sending"
		return dtdal.Email.UpdateEmail(c, tx, email)
	}, nil); err != nil {
		err = fmt.Errorf("failed to update email status to 'queued': %w", err)
		if dal.IsNotFound(err) {
			log.Warningf(c, err.Error())
			return nil // Do not retry
		} else if errors.Is(err, ErrEmailIsInWrongStatus) {
			log.Warningf(c, err.Error())
			return nil // Do not retry
		}
		log.Errorf(c, err.Error())
		return err // Retry
	}

	var awsSesMessageID string
	if awsSesMessageID, err = SendEmail(c, email.From, email.To, email.Subject, email.BodyText, email.BodyHtml); err != nil {
		log.Errorf(c, "Failed to send email: %v", err)

		if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
			if email, err = dtdal.Email.GetEmailByID(c, id); err != nil {
				return err
			}
			if email.Status != "sending" {
				return fmt.Errorf("%w: expected 'sending' got email.Status=%s", ErrEmailIsInWrongStatus, email.Status)
			}
			email.Status = "error"
			email.Error = err.Error()
			return dtdal.Email.UpdateEmail(c, tx, email)
		}); err != nil {
			log.Errorf(c, err.Error())
		}
		return nil // Do not retry
	}

	log.Infof(c, "Sent email, message ID: %v", awsSesMessageID)

	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		if email, err = dtdal.Email.GetEmailByID(c, id); err != nil {
			return err
		}
		if email.Status != "sending" {
			return fmt.Errorf("%w: expected 'sending' got email.Status=%s", ErrEmailIsInWrongStatus, email.Status)
		}
		email.Status = "sent"
		email.DtSent = time.Now()
		email.AwsSesMessageID = awsSesMessageID
		return dtdal.Email.UpdateEmail(c, tx, email)
	}); err != nil {
		log.Errorf(c, err.Error())
		err = nil // Do not retry!
	}
	return nil // Do not retry!
}
