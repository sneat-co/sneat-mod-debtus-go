package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/api/dto"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/facade"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"github.com/strongo/db"
	"github.com/strongo/decimal"
	"github.com/strongo/log"
)

func handleGetTransfer(c context.Context, w http.ResponseWriter, r *http.Request) {
	if transferID := getID(c, w, r, "id"); transferID == 0 {
		return
	} else {
		transfer, err := dal.Transfer.GetTransferByID(c, transferID)
		if hasError(c, w, err, models.TransferKind, transferID, http.StatusBadRequest) {
			return
		}

		if transfer, err = facade.CheckTransferCreatorNameAndFixIfNeeded(c, w, transfer); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		record := NewReceiptTransferDto(c, transfer)

		jsonToResponse(c, w, &record)
	}
}

type transferSourceSetToAPI struct {
	appPlatform string
	createdOnID string
}

func (s transferSourceSetToAPI) PopulateTransfer(t *models.TransferEntity) {
	t.CreatedOnPlatform = s.appPlatform
	t.CreatedOnID = s.createdOnID
}

func handleCreateTransfer(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	var direction models.TransferDirection
	switch r.PostFormValue("direction") {
	case "contact-to-user":
		direction = models.TransferDirectionCounterparty2User
	case "user-to-contact":
		direction = models.TransferDirectionUser2Counterparty
	default:
		w.WriteHeader(http.StatusBadRequest)
		m := "Unknown direction: " + r.PostFormValue("direction")
		log.Debugf(c, m)
		w.Write([]byte(m))
		return
	}
	amountValue, err := decimal.ParseDecimal64p2(r.PostFormValue("amount"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if amountValue < 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("amount < 0"))
		return
	}
	currency := r.PostFormValue("currency")
	if len(currency) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("len(currency) > 30"))
		return
	}

	amountWithCurrency := models.NewAmount(models.Currency(currency), amountValue)

	contactID, err := strconv.ParseInt(r.PostFormValue("contactID"), 10, 64)
	if err != nil {
		BadRequestError(c, w, err)
		return
	}

	var (
		returnToTransferID int64
		dueOn              time.Time
	)

	isReturn := r.PostFormValue("isReturn") == "true"

	if dueStr := r.PostFormValue("due"); dueStr != "" {
		if dueOn, err = time.Parse("2006-01-02", dueStr); err != nil {
			BadRequestError(c, w, err)
			return
		}
	}

	//user, err := dal.User.GetUserByID(c, authInfo.AppUserIntID)
	//if err != nil {
	//	hashedWriter.WriteHeader(http.StatusInternalServerError)
	//	hashedWriter.Write([]byte(errors.Wrap(err, "Failed to get user")))
	//}
	if isReturn {
		if counterparty, err := dal.Contact.GetContactByID(c, contactID); err != nil {
			if db.IsNotFound(err) {
				BadRequestError(c, w, err)
			} else {
				InternalError(c, w, err)
			}
			return
		} else {
			balance := counterparty.Balance()
			if balanceAmount, ok := balance[amountWithCurrency.Currency]; !ok {
				BadRequestMessage(c, w, fmt.Sprintf("No balance for %v", amountWithCurrency.Currency))
			} else {
				switch direction {
				case models.TransferDirectionUser2Counterparty:
					if balanceAmount > 0 {
						BadRequestMessage(c, w, fmt.Sprintf("balanceAmount > 0 && direction == %v", direction))
					}
				case models.TransferDirectionCounterparty2User:
					if balanceAmount < 0 {
						BadRequestMessage(c, w, fmt.Sprintf("balanceAmount < 0 && direction == %v", direction))
					}
				}
			}
		}
	}

	var from, to *models.TransferCounterpartyInfo

	switch direction {
	case models.TransferDirectionUser2Counterparty:
		from = models.NewFrom(authInfo.UserID, r.PostFormValue("comment"))
		to = models.NewTo(contactID)
	case models.TransferDirectionCounterparty2User:
		from = models.NewTo(contactID)
		to = models.NewFrom(authInfo.UserID, r.PostFormValue("comment"))
	default:
		panic(fmt.Sprintf("Unknown direction: %v", direction))
	}

	platform := r.PostFormValue("platform")
	if len(platform) > 20 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("len(platform) > 20"))
	}
	switch platform {
	case "web":
	case "ios":
	case "android":
	default:
		BadRequestMessage(c, w, "Unknown platform: "+platform)
		return
	}

	var appUser models.AppUser
	if appUser, err = dal.User.GetUserByID(c, authInfo.UserID); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	newTransfer := facade.NewTransferInput(getEnvironment(r),
		transferSourceSetToAPI{appPlatform: platform, createdOnID: r.Host},
		appUser,
		"",
		isReturn, returnToTransferID,
		from, to,
		amountWithCurrency, dueOn, models.TransferInterest{})
	output, err := facade.Transfers.CreateTransfer(c, newTransfer)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	//userBalance := json.RawMessage(user.BalanceJson)
	log.Infof(c, "transfer.DtDueOn: %v", output.Transfer.DtDueOn)
	response := dto.CreateTransferResponse{
		Transfer: dto.TransferToDto(authInfo.UserID, output.Transfer),
	}

	var counterparty models.Contact
	switch output.Transfer.CreatorUserID {
	case output.Transfer.From().UserID:
		counterparty = output.To.Contact
	case output.Transfer.To().UserID:
		counterparty = output.From.Contact
	default:
		panic("Unknown direction")
	}
	if counterparty.BalanceJson != "" {
		counterpartyBalance := json.RawMessage(counterparty.BalanceJson)
		response.CounterpartyBalance = &counterpartyBalance
	}
	jsonToResponse(c, w, response)
}
