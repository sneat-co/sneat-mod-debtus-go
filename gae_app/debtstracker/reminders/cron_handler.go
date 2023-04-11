package reminders

import (
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal/gaedal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/app/gae"
	"github.com/strongo/log"
	"google.golang.org/appengine/v2/datastore"
)

func CronSendReminders(c context.Context, w http.ResponseWriter, r *http.Request) {
	query := datastore.NewQuery(models.ReminderKind).
		Filter("Status =", models.ReminderStatusCreated).
		Filter("DtNext >", time.Time{}).Filter("DtNext <", time.Now()).Order("DtNext").
		Limit(100).KeysOnly()
	//KeysOnly()
	reminderKeys, err := query.GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "Failed to load due transfers: %v", err)
		return
	}

	if len(reminderKeys) == 0 {
		log.Debugf(c, "No reminders to send")
		return
	}

	log.Debugf(c, "Loaded %d reminder(s)", len(reminderKeys))

	for _, reminderKey := range reminderKeys {
		reminderID := reminderKey.IntID()
		task := gaedal.CreateSendReminderTask(c, reminderID)
		task.Name = fmt.Sprintf("r_%v_%v", reminderID, time.Now().Format("200601021504"))
		if _, err := gae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			log.Errorf(c, "Failed to add delayed task for reminder %d", reminderKey.IntID())
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
