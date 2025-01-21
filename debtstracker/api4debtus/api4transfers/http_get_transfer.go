package api4transfers

import (
	"context"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-core-modules/common4all"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/api4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"net/http"
)

func HandleGetTransfer(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if transferID := common4all.GetStrID(ctx, w, r, "id"); transferID == "" {
		return
	} else {
		transfer, err := facade4debtus.Transfers.GetTransferByID(ctx, nil, transferID)
		if common4all.HasError(ctx, w, err, models4debtus.TransfersCollection, transferID, http.StatusBadRequest) {
			return
		}

		err = facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
			if err = facade4debtus.CheckTransferCreatorNameAndFixIfNeeded(ctx, tx, transfer); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			return err
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		record := api4debtus.NewReceiptTransferDto(ctx, transfer)

		common4all.JsonToResponse(ctx, w, &record)
	}
}

type transferSourceSetToAPI struct {
	appPlatform string
	createdOnID string
}

func (s transferSourceSetToAPI) PopulateTransfer(t *models4debtus.TransferData) {
	t.CreatedOnPlatform = s.appPlatform
	t.CreatedOnID = s.createdOnID
}
