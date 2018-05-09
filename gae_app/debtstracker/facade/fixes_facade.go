package facade

import (
	"fmt"
	"net/http"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/log"
)

func CheckTransferCreatorNameAndFixIfNeeded(c context.Context, w http.ResponseWriter, transfer models.Transfer) (models.Transfer, error) {
	if transfer.Creator().UserName == "" {
		user, err := User.GetUserByID(c, transfer.CreatorUserID)
		if err != nil {
			return transfer, err
		}

		creatorFullName := user.FullName()
		if creatorFullName == "" || creatorFullName == models.NoName {
			log.Debugf(c, "Can't fix transfers creator name as user entity has no name defined.")
			return transfer, nil
		}

		logMessage := fmt.Sprintf("Fixing transfer(%d).Creator().UserName, created: %v", transfer.ID, transfer.DtCreated)
		if transfer.DtCreated.After(time.Date(2017, 8, 1, 0, 0, 0, 0, time.UTC)) {
			log.Warningf(c, logMessage)
		} else {
			log.Infof(c, logMessage)
		}

		if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
			if transfer, err = GetTransferByID(c, transfer.ID); err != nil {
				return err
			}
			if transfer.Creator().UserName == "" {
				changed := false
				switch transfer.Direction() {
				case models.TransferDirectionUser2Counterparty:
					transfer.From().UserName = creatorFullName
					changed = true
				case models.TransferDirectionCounterparty2User:
					transfer.To().UserName = creatorFullName
					changed = true
				}
				if changed {
					return Transfers.SaveTransfer(c, transfer)
				}
			}
			return nil
		}, nil); err != nil {
			return transfer, err
		}
	}
	return transfer, nil
}
