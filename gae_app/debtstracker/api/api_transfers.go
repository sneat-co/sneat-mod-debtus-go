package api

//go:generate ffjson $GOFILE

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"golang.org/x/net/context"
	"net/http"
	"time"
	"github.com/pquerna/ffjson/ffjson"
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

type TransferDto struct {
	Id            int64
	Created       time.Time
	Amount        models.Amount
	IsReturn      bool
	CreatorUserID int64
	From          *ContactDto
	To            *ContactDto
	Due           time.Time `json:",omitempty"`
}

func (t TransferDto) String() string {
	if b, err := ffjson.Marshal(t); err != nil {
		return err.Error()
	} else {
		return string(b)
	}
}

type TransfersResultDto struct {
	HasMoreTransfers bool `json:",omitempty"`
	Transfers        []*TransferDto `json:",omitempty"`
}

func transfersToDto(userID int64, transfers []models.Transfer) []*TransferDto {
	transfersDto := make([]*TransferDto, len(transfers))
	for i, transfer := range transfers {
		transfersDto[i] = transferToDto(userID, transfer)
	}
	return transfersDto
}

func transferToDto(userID int64, transfer models.Transfer) *TransferDto {
	transferDto := TransferDto{
		Id:            transfer.ID,
		Amount:        transfer.GetAmount(),
		Created:       transfer.DtCreated,
		CreatorUserID: transfer.CreatorUserID,
		IsReturn:      transfer.IsReturn,
		Due:           transfer.DtDueOn,
	}

	from := NewContactDto(*transfer.From())
	to := NewContactDto(*transfer.To())

	switch userID {
	case 0:
		transferDto.From = &from
		transferDto.To = &to
	case from.UserID:
		transferDto.To = &to
	case to.UserID:
		transferDto.From = &from
	default:
		transferDto.From = &from
		transferDto.To = &to
	}

	return &transferDto
}

func transfersToResponse(c context.Context, w http.ResponseWriter, userID int64, transfers []models.Transfer, hasMore bool) {
	jsonToResponse(c, w, TransfersResultDto{
		Transfers:        transfersToDto(userID, transfers),
		HasMoreTransfers: hasMore,
	})
}
