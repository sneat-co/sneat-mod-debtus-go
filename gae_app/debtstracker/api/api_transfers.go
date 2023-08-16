package api

import (
	"net/http"

	"context"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/api/dto"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/auth"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/models"
)

func handleAdminLatestTransfers(c context.Context, w http.ResponseWriter, r *http.Request, _ auth.AuthInfo) {
	transfers, err := dtdal.Transfer.LoadLatestTransfers(c, 0, 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(([]byte)(err.Error()))
	}
	transfersToResponse(c, w, "", transfers, true)
}

func handleUserTransfers(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser) {
	transfers, hasMore, err := dtdal.Transfer.LoadTransfersByUserID(c, user.ID, 0, 100)
	if hasError(c, w, err, "", "", http.StatusInternalServerError) {
		return
	}
	transfersToResponse(c, w, user.ID, transfers, hasMore)
}

func transfersToResponse(c context.Context, w http.ResponseWriter, userID string, transfers []models.Transfer, hasMore bool) {
	jsonToResponse(c, w, dto.TransfersResultDto{
		Transfers:        dto.TransfersToDto(userID, transfers),
		HasMoreTransfers: hasMore,
	})
}
