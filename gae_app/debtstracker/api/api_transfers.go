package api



import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"net/http"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/api/dto"
)

func handleAdminLatestTransfers(c context.Context, w http.ResponseWriter, r *http.Request, _ auth.AuthInfo) {
	transfers, err := dal.Transfer.LoadLatestTransfers(c, 0, 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(([]byte)(err.Error()))
	}
	transfersToResponse(c, w, 0, transfers, true)
}

func handleUserTransfers(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser) {
	transfers, hasMore, err := dal.Transfer.LoadTransfersByUserID(c, user.ID, 0, 100)
	if hasError(c, w, err, "", 0, http.StatusInternalServerError) {
		return
	}
	transfersToResponse(c, w, user.ID, transfers, hasMore)
}


func transfersToResponse(c context.Context, w http.ResponseWriter, userID int64, transfers []models.Transfer, hasMore bool) {
	jsonToResponse(c, w, dto.TransfersResultDto{
		Transfers:        dto.TransfersToDto(userID, transfers),
		HasMoreTransfers: hasMore,
	})
}
