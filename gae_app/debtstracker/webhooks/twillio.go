package webhooks

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"github.com/qedus/nds"
	"github.com/strongo/app/log"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
	"time"
)

func TwilioWebhook(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := r.ParseForm()
	if err != nil {
		log.Errorf(c, "Failed to parse POST form: %v", err)
		return
	}
	log.Infof(c, "BODY: %v", r.Form)
	smsSid := r.PostFormValue("SmsSid")
	messageStatus := r.PostFormValue("MessageStatus")

	err = dal.DB.RunInTransaction(c, func(tc context.Context) error {
		var smsEntity models.TwilioSmsEntity
		smsEntityKey := datastore.NewKey(tc, models.TwilioSmsKind, smsSid, 0, nil)
		err := nds.Get(tc, smsEntityKey, &smsEntity)
		if err != nil {
			return err
		}
		if smsEntity.Status != messageStatus {
			smsEntity.Status = messageStatus
			switch messageStatus {
			case "sent":
				smsEntity.DtSent = time.Now()
			case "delivered":
				smsEntity.DtDelivered = time.Now()
			}
			nds.Put(tc, smsEntityKey, &smsEntity)
		}
		return nil
	}, nil)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			log.Infof(c, "Unknown SMS: %v", smsSid)
		} else {
			log.Errorf(c, "Failed to process SMS update: %v", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Infof(c, "Success")
		w.WriteHeader(http.StatusOK)
	}
}
