package facade

import (
	"fmt"
	"net/http"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/log"
	"golang.org/x/net/context"
)

func CheckTransferCreatorNameAndFixIfNeeded(c context.Context, w http.ResponseWriter, transfer models.Transfer) (models.Transfer, error) {
	if transfer.Creator().ContactName == "" {
		user, err := dal.User.GetUserByID(c, transfer.CreatorUserID)
		if err != nil {
			return transfer, err
		}

		creatorFullName := user.FullName()
		if creatorFullName == "" || creatorFullName == models.NO_NAME {
			return transfer, nil
		}

		logMessage := fmt.Sprintf("Fixing transfer(%d).Creator().ContactName, created: %v", transfer.ID, transfer.DtCreated)
		if transfer.DtCreated.After(time.Date(2017, 8, 1, 0, 0, 0, 0, time.UTC)) {
			log.Warningf(c, logMessage)
		} else {
			log.Infof(c, logMessage)
		}

		if err = dal.DB.RunInTransaction(c, func(c context.Context) error {
			if transfer, err = dal.Transfer.GetTransferByID(c, transfer.ID); err != nil {
				return err
			}
			if transfer.Creator().ContactName == "" {
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
					return dal.Transfer.SaveTransfer(c, transfer)
				}
			}
			return nil
		}, nil); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		return transfer, err
	}
	return transfer, nil
}
