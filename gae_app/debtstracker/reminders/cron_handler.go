package reminders

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"reflect"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/dtdal/gaedal"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/app/gae"
	"github.com/strongo/log"
)

func CronSendReminders(c context.Context, w http.ResponseWriter, r *http.Request) {
	query := dal.From(models.ReminderKind).
		WhereField("Status", dal.Equal, models.ReminderStatusCreated).
		WhereField("DtNext", dal.GreaterThen, time.Time{}).
		WhereField("DtNext", dal.LessThen, time.Now()).
		OrderBy(dal.AscendingField("DtNext")).
		SelectKeysOnly(reflect.Int)
	query.Limit = 100

	db, err := facade.GetDatabase(c)
	if err != nil {
		log.Errorf(c, "Failed to get database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var reader dal.Reader
	if reader, err = db.QueryReader(c, query); err != nil {
		log.Errorf(c, "Failed to load due transfers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var reminderIDs []int
	if reminderIDs, err = dal.SelectAllIDs[int](reader, query.Limit); err != nil {
		log.Errorf(c, "Failed to load due transfers: %v", err)
		return
	}

	if len(reminderIDs) == 0 {
		log.Debugf(c, "No reminders to send")
		return
	}

	log.Debugf(c, "Loaded %d reminder(s)", len(reminderIDs))

	for _, reminderID := range reminderIDs {
		task := gaedal.CreateSendReminderTask(c, reminderID)
		task.Name = fmt.Sprintf("r_%d_%v", reminderID, time.Now().Format("200601021504"))
		if _, err := gae.AddTaskToQueue(c, task, common.QUEUE_REMINDERS); err != nil {
			log.Errorf(c, "Failed to add delayed task for reminder %d", reminderID)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
