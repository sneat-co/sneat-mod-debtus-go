package api4transfers

import (
	"context"
	"github.com/sneat-co/sneat-core-modules/auth/token4auth"
	"github.com/sneat-co/sneat-core-modules/common4all"
	"github.com/sneat-co/sneat-core-modules/userus/dal4userus"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus/dto4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"net/http"
)

func HandleCreateTransfer(ctx context.Context, w http.ResponseWriter, r *http.Request, authInfo token4auth.AuthInfo) {
	var request facade4debtus.CreateTransferRequest
	apicore.HandleAuthenticatedRequestWithBody(w, r, &request, verify.DefaultJsonWithAuthRequired, http.StatusCreated,
		func(ctx context.Context, userCtx facade.UserContext) (interface{}, error) {
			var from, to *models4debtus.TransferCounterpartyInfo

			appUser, err := dal4userus.GetUserByID(ctx, nil, authInfo.UserID)
			if err != nil {
				return nil, err
			}

			newTransfer := facade4debtus.NewTransferInput(common4all.GetEnvironment(r),
				transferSourceSetToAPI{appPlatform: "api4debtus", createdOnID: r.Host},
				appUser,
				request,
				from, to,
			)

			output, err := facade4debtus.Transfers.CreateTransfer(ctx, newTransfer)
			if err != nil {
				return nil, err
			}

			response := dto4debtus.CreateTransferResponse{
				Transfer: dto4debtus.TransferToDto(authInfo.UserID, output.Transfer),
			}

			var counterparty models4debtus.DebtusSpaceContactEntry
			switch output.Transfer.Data.CreatorUserID {
			case output.Transfer.Data.From().UserID:
				counterparty = output.To.DebtusContact
			case output.Transfer.Data.To().UserID:
				counterparty = output.From.DebtusContact
			default:
				panic("Unknown direction")
			}
			if len(counterparty.Data.Balance) > 0 {
				response.CounterpartyBalance = counterparty.Data.Balance
			}
			return response, err
		})
}
