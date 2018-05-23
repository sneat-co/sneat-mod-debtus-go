package sms

import (
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"context"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/app"
	"github.com/strongo/gotwilio"
	"github.com/strongo/log"
	"google.golang.org/appengine/urlfetch"
)

func SendSms(c context.Context, isLive bool, toPhoneNumber, smsText string) (isTestSender bool, smsResponse *gotwilio.SmsResponse, twilioException *gotwilio.Exception, err error) {
	var (
		accountSid   string
		accountToken string
		fromNumber   string
	)

	if isLive {
		accountSid = common.TWILIO_LIVE_ACCOUNT_SID
		accountToken = common.TWILIO_LIVE_ACCOUNT_TOKEN
		fromNumber = common.TWILIO_LIVE_FROM_US
	} else {
		accountSid = common.TWILIO_TEST_ACCOUNT_SID
		accountToken = common.TWILIO_TEST_ACCOUNT_TOKEN
		fromNumber = common.TWILIO_TEST_FROM
	}

	twilio := gotwilio.NewTwilioClientCustomHTTP(accountSid, accountToken, urlfetch.Client(c))

	if smsResponse, twilioException, err = twilio.SendSMS(
		fromNumber,
		toPhoneNumber,
		smsText,
		"https://debtstracker-io.appspot.com/webooks/twilio/sms/status?sender=callback-url",
		common.TWILIO_APPLICATION_SID,
	); err != nil {
		return
	}

	if twilioException != nil && twilioException.Code == 21211 && len(toPhoneNumber) == 12 && strings.HasPrefix(toPhoneNumber, "+8") { // is not a valid phone number
		correctedPhoneNumber := strings.Replace(toPhoneNumber, "+8", "+7", 1)
		log.Warningf(c, "%v. Will try to send after changing phone number from %v to %v", twilioException.Message, toPhoneNumber, correctedPhoneNumber)
		smsResponse, twilioException, err = twilio.SendSMS(
			fromNumber,
			correctedPhoneNumber,
			smsText,
			"https://debtstracker-io.appspot.com/webooks/twilio/sms/status?sender=callback-url",
			common.TWILIO_APPLICATION_SID,
		)
	}
	return
}

func TwilioExceptionToMessage(ec strongo.ExecutionContext, ex *gotwilio.Exception) (messageText string, tryAnotherNumber bool) {
	switch ex.Code {
	case 21211: // Is not a valid phone number. https://www.twilio.com/docs/errors/21211
		tryAnotherNumber = true
		messageText = ec.Translate(trans.MESSAGE_TEXT_INVALID_PHONE_NUMBER)
	case 21614: // Is is not a mobile number https://www.twilio.com/docs/errors/21614}
		tryAnotherNumber = true
		messageText = ec.Translate(trans.MESSAGE_TEXT_PHONE_NUMBER_IS_NOT_SMS_CAPABLE)
	case 21612: // is not currently reachable using the 'From' phone number via SMS. https://www.twilio.com/docs/errors/21612
		tryAnotherNumber = true
		messageText = ec.Translate("is not currently reachable using the 'From' phone number via SMS")
	case 21408: // Permission to send an SMS has not been enabled for the region indicated by the 'To' number: https://www.twilio.com/docs/errors/21408
		tryAnotherNumber = true
		messageText = ec.Translate("Permission to send an SMS has not been enabled for the region indicated by the 'To' number")
	case 21610: // The message From/To pair violates a blacklist rule. https://www.twilio.com/docs/errors/21610
		tryAnotherNumber = true
		messageText = ec.Translate("The message From/To pair violates a blacklist rule.")
	}
	return
}
