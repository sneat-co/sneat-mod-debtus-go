package gaedal

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/pkg/errors"
	"github.com/strongo/db"
	"github.com/strongo/db/gaedb"
	"github.com/strongo/decimal"
	"github.com/strongo/gotwilio"
	"github.com/strongo/log"
	"context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type TwilioDalGae struct {
}

func NewTwilioDalGae() TwilioDalGae {
	return TwilioDalGae{}
}

func (_ TwilioDalGae) GetLastTwilioSmsesForUser(c context.Context, userID int64, to string, limit int) (result []models.TwilioSms, err error) {
	query := datastore.NewQuery(models.TwilioSmsKind).Filter("UserID =", userID).Order("-DtCreated").Limit(limit)
	if to != "" {
		query = query.Filter("To =", to)
	}
	entities := make([]*models.TwilioSmsEntity, 0, 1)
	keys := make([]*datastore.Key, 0, limit)
	if keys, err = query.GetAll(c, &entities); err != nil || len(keys) == 0 {
		return
	}
	result = make([]models.TwilioSms, len(keys))
	for i, entity := range entities {
		result[i] = models.TwilioSms{StringID: db.StringID{ID: keys[i].StringID()}, TwilioSmsEntity: entity}
	}
	return
}

func (_ TwilioDalGae) SaveTwilioSms(
	c context.Context,
	smsResponse *gotwilio.SmsResponse,
	transfer models.Transfer,
	phoneContact models.PhoneContact,
	userID int64,
	tgChatID int64,
	smsStatusMessageID int,
) (twilioSms models.TwilioSms, err error) {
	var twilioSmsEntity models.TwilioSmsEntity
	if err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
		userKey := NewAppUserKey(c, userID)
		transferKey := NewTransferKey(tc, transfer.ID)
		counterpartyKey := NewContactKey(tc, transfer.Counterparty().ContactID)
		twilioSmsKey := gaedb.NewKey(tc, models.TwilioSmsKind, smsResponse.Sid, 0, nil)
		var (
			appUserEntity      models.AppUserEntity
			counterpartyEntity models.ContactEntity
		)
		if err := gaedb.GetMulti(tc, []*datastore.Key{userKey, twilioSmsKey, transferKey, counterpartyKey}, []interface{}{&appUserEntity, &twilioSmsEntity, transfer.TransferEntity, &counterpartyEntity}); err != nil {
			if multiError, ok := err.(appengine.MultiError); ok {
				if multiError[1] == datastore.ErrNoSuchEntity {
					twilioSmsEntity = models.NewTwilioSmsFromSmsResponse(userID, smsResponse)
					twilioSmsEntity.CreatorTgChatID = tgChatID
					twilioSmsEntity.CreatorTgSmsStatusMessageID = smsStatusMessageID

					appUserEntity.SmsCount += 1
					transfer.SmsCount += 1

					appUserEntity.SmsCost += twilioSmsEntity.Price
					transfer.SmsCost += twilioSmsEntity.Price

					smsPriceUSD := decimal.NewDecimal64p2FromFloat64(float64(twilioSmsEntity.Price))
					twilioSmsEntity.PriceUSD = smsPriceUSD
					appUserEntity.SmsCostUSD += smsPriceUSD
					transfer.SmsCostUSD += smsPriceUSD

					keysToPut := []*datastore.Key{
						userKey,
						twilioSmsKey,
						transferKey,
					}
					entitiesToPut := []interface{}{
						&appUserEntity,
						&twilioSmsEntity,
						transfer.TransferEntity,
					}
					if counterpartyEntity.PhoneContact.PhoneNumber != phoneContact.PhoneNumber {
						counterpartyEntity.PhoneContact = phoneContact
						keysToPut = append(keysToPut, counterpartyKey)
						entitiesToPut = append(entitiesToPut, &counterpartyEntity)
					}
					if _, err = gaedb.PutMulti(tc, keysToPut, entitiesToPut); err != nil {
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
	}, dal.CrossGroupTransaction); err != nil {
		err = errors.Wrap(err, "Failed to save Twilio response to DB")
		return
	}
	twilioSms = models.TwilioSms{StringID: db.StringID{ID: smsResponse.Sid}, TwilioSmsEntity: &twilioSmsEntity}
	return
}
