package facade

import (
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"net/http"
	"time"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/log"
)

func CheckTransferCreatorNameAndFixIfNeeded(c context.Context, w http.ResponseWriter, transfer models.Transfer) (models.Transfer, error) {
	if transfer.Data.Creator().UserName == "" {
		user, err := User.GetUserByID(c, transfer.Data.CreatorUserID)
		if err != nil {
			return transfer, err
		}

		creatorFullName := user.Data.FullName()
		if creatorFullName == "" || creatorFullName == models.NoName {
			log.Debugf(c, "Can't fix transfers creator name as user entity has no name defined.")
			return transfer, nil
		}

		logMessage := fmt.Sprintf("Fixing transfer(%d).Creator().UserName, created: %v", transfer.ID, transfer.Data.DtCreated)
		if transfer.Data.DtCreated.After(time.Date(2017, 8, 1, 0, 0, 0, 0, time.UTC)) {
			log.Warningf(c, logMessage)
		} else {
			log.Infof(c, logMessage)
		}

		var db dal.Database
		if db, err = GetDatabase(c); err != nil {
			return transfer, err
		}

		if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
			if transfer, err = Transfers.GetTransferByID(c, tx, transfer.ID); err != nil {
				return err
			}
			if transfer.Data.Creator().UserName == "" {
				changed := false
				switch transfer.Data.Direction() {
				case models.TransferDirectionUser2Counterparty:
					transfer.Data.From().UserName = creatorFullName
					changed = true
				case models.TransferDirectionCounterparty2User:
					transfer.Data.To().UserName = creatorFullName
					changed = true
				}
				if changed {
					return Transfers.SaveTransfer(c, tx, transfer)
				}
			}
			return nil
		}, nil); err != nil {
			return transfer, err
		}
	}
	return transfer, nil
}
