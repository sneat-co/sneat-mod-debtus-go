package gaedal

import (
	"context"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"github.com/strongo/decimal"
	"github.com/strongo/gotwilio"
	"github.com/strongo/log"
	"google.golang.org/appengine"
)

type TwilioDalGae struct {
}

func NewTwilioDalGae() TwilioDalGae {
	return TwilioDalGae{}
}

func (TwilioDalGae) GetLastTwilioSmsesForUser(c context.Context, tx dal.ReadSession, userID int64, to string, limit int) (result []models.TwilioSms, err error) {
	q := dal.From(models.TwilioSmsKind).
		WhereField("UserID", dal.Equal, userID).
		OrderBy(dal.DescendingField("DtCreated"))

	if to != "" {
		q = q.WhereField("To", dal.Equal, to)
	}
	query := q.Limit(limit).SelectInto(models.NewTwilioSmsRecord)
	var records []dal.Record
	if records, err = tx.QueryAllRecords(c, query); err != nil {
		return
	}
	result = models.NewTwilioSmsFromRecords(records)
	return
}

func (TwilioDalGae) SaveTwilioSms(
	c context.Context,
	smsResponse *gotwilio.SmsResponse,
	transfer models.Transfer,
	phoneContact models.PhoneContact,
	userID int64,
	tgChatID int64,
	smsStatusMessageID int,
) (twilioSms models.TwilioSms, err error) {
	var twilioSmsEntity models.TwilioSmsData
	var db dal.Database
	if db, err = GetDatabase(c); err != nil {
		return
	}
	if err = db.RunReadwriteTransaction(c, func(tc context.Context, tx dal.ReadwriteTransaction) error {
		user := models.NewAppUser(userID, nil)
		twilioSms = models.NewTwilioSms(smsResponse.Sid, nil)
		counterparty := models.NewContact(transfer.Data.Counterparty().ContactID, nil)
		if err := tx.GetMulti(tc, []dal.Record{user.Record, twilioSms.Record, transfer.Record, counterparty.Record}); err != nil {
			if multiError, ok := err.(appengine.MultiError); ok {
				if multiError[1] == dal.ErrNoMoreRecords {
					twilioSmsEntity = models.NewTwilioSmsFromSmsResponse(userID, smsResponse)
					twilioSmsEntity.CreatorTgChatID = tgChatID
					twilioSmsEntity.CreatorTgSmsStatusMessageID = smsStatusMessageID

					user.Data.SmsCount += 1
					transfer.Data.SmsCount += 1

					user.Data.SmsCost += twilioSmsEntity.Price
					transfer.Data.SmsCost += twilioSmsEntity.Price

					smsPriceUSD := decimal.NewDecimal64p2FromFloat64(float64(twilioSmsEntity.Price))
					twilioSmsEntity.PriceUSD = smsPriceUSD
					user.Data.SmsCostUSD += smsPriceUSD
					transfer.Data.SmsCostUSD += smsPriceUSD

					recordsToPut := []dal.Record{
						user.Record,
						twilioSms.Record,
						transfer.Record,
					}
					if counterparty.Data.PhoneContact.PhoneNumber != phoneContact.PhoneNumber {
						counterparty.Data.PhoneContact = phoneContact
						recordsToPut = append(recordsToPut, counterparty.Record)
					}
					if err = tx.SetMulti(tc, recordsToPut); err != nil {
						log.Errorf(c, "Failed to save Twilio SMS")
						return err
					}
					return err
				} else if multiError[1] == nil {
					log.Warningf(c, "Twillio SMS already saved to DB (1)")
				}
			} else {
				return err
			}
		} else {
			log.Warningf(c, "Twillio SMS already saved to DB (2)")
		}
		return nil
	}); err != nil {
		err = fmt.Errorf("failed to save Twilio response to DB: %w", err)
		return
	}
	return
}
