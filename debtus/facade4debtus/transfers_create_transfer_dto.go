package facade4debtus

import (
	"errors"
	"fmt"
	"github.com/crediterra/money"
	"github.com/sneat-co/sneat-core-modules/contactus/dal4contactus"
	"github.com/sneat-co/sneat-core-modules/spaceus/dto4spaceus"
	"github.com/sneat-co/sneat-core-modules/userus/dbo4userus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/strongo/validation"
	"time"
)

type CreateTransferRequest struct {
	dto4spaceus.SpaceRequest
	Direction          models4debtus.TransferDirection `json:"direction"`
	Amount             money.Amount                    `json:"amount"`
	FromContactID      string                          `json:"fromContactID"`
	ToContactID        string                          `json:"toContactID"`
	BillID             string                          `json:"billID"`
	IsReturn           bool                            `json:"isReturn,omitempty"`
	ReturnToTransferID string                          `json:"returnToTransferID,omitempty"`
	DueOn              *time.Time                      `json:"dueOn,omitempty"`
	Interest           *models4debtus.TransferInterest `json:"interest,omitempty"`
}

var (
	ErrTransferAmountCannotBeNegative = validation.NewErrBadRequestFieldValue("amount.value", "transfer amount cannot be negative")
	ErrTransferDirectionIsInvalid     = validation.NewErrBadRequestFieldValue("direction", "unknown transfer direction")
)

func (v *CreateTransferRequest) Validate() error {
	if err := v.SpaceRequest.Validate(); err != nil {
		return err
	}
	switch v.Direction {
	case models4debtus.TransferDirectionCounterparty2User:
		if v.FromContactID == "" {
			return validation.NewErrRequestIsMissingRequiredField("fromContactID")
		}
	case models4debtus.TransferDirectionUser2Counterparty:
		if v.ToContactID == "" {
			return validation.NewErrRequestIsMissingRequiredField("toContactID")
		}
	case models4debtus.TransferDirection3dParty:
		if v.FromContactID == "" {
			return validation.NewErrRequestIsMissingRequiredField("fromContactID")
		}
		if v.ToContactID == "" {
			return validation.NewErrRequestIsMissingRequiredField("toContactID")
		}
	}
	if !models4debtus.IsKnownTransferDirection(v.Direction) {
		return fmt.Errorf("%w: %v", ErrTransferDirectionIsInvalid, v.Direction)
	}
	if v.Amount.Value < 0 {
		return ErrTransferAmountCannotBeNegative
	}
	if v.Amount.Currency == "" {
		return validation.NewErrRequestIsMissingRequiredField("amount.currency")
	}
	return nil
}

type CreateTransferInput struct {
	Env         string // TODO: I believe we don't need this
	Source      dtdal.TransferSource
	CreatorUser dbo4userus.UserEntry
	Request     CreateTransferRequest
	From, To    *models4debtus.TransferCounterpartyInfo
}

type ParticipantEntries struct { // Consider making it non exported
	User          dbo4userus.UserEntry
	SpaceID       string
	Contact       dal4contactus.ContactEntry
	DebtusSpace   models4debtus.DebtusSpaceEntry
	DebtusContact models4debtus.DebtusSpaceContactEntry
}

type CreateTransferOutput struct {
	Transfer          models4debtus.TransferEntry
	ReturnedTransfers []models4debtus.TransferEntry
	From, To          *ParticipantEntries
}

func (input CreateTransferInput) Direction() (direction models4debtus.TransferDirection) {
	if input.CreatorUser.ID == "" {
		panic("CreateTransferInput.CreatorUserID == 0")
	}
	switch input.CreatorUser.ID {
	case input.From.UserID:
		return models4debtus.TransferDirectionUser2Counterparty
	case input.To.UserID:
		return models4debtus.TransferDirectionCounterparty2User
	default:
		if input.Request.BillID == "" {
			panic("Not able to detect direction")
		}
		return models4debtus.TransferDirection3dParty
	}
}

func (input CreateTransferInput) CreatorContactID() string {
	switch input.CreatorUser.ID {
	case input.From.UserID:
		return input.To.ContactID
	case input.To.UserID:
		return input.From.ContactID
	}
	panic("Can't get creator's contact ContactID as it's a 3d-party transfer")
}

func (output CreateTransferOutput) Validate() {
	if output.Transfer.ID == "" {
		panic("TransferEntry.ContactID == 0")
	}
	if output.Transfer.Data == nil {
		panic("TransferData == nil")
	}
}

func (input CreateTransferInput) Validate() error {
	if input.Source == nil {
		return errors.New("source == nil")
	}
	if input.CreatorUser.ID == "" {
		return errors.New("creatorUser.ContactID == 0")
	}
	if input.CreatorUser.Data == nil {
		return errors.New("creatorUser.DebutsAppUserDataOBSOLETE == nil")
	}
	if err := input.Request.Validate(); err != nil {
		return err
	}
	if input.Request.Amount.Value <= 0 {
		return errors.New("amount.Value <= 0")
	}
	if input.From == nil {
		return errors.New("from == nil")
	}
	if input.To == nil {
		return errors.New("to == nil")
	}

	if (input.From.ContactID == "" && input.To.ContactID == "") || (input.From.UserID == "" && input.To.UserID == "") {
		return errors.New("(from.ContactID == 0  && to.ContactID == 0) || (from.UserID == 0 && to.UserID == 0)")
	}
	if input.From.UserID != "" && input.To.ContactID == "" && input.To.UserID == "" {
		return errors.New("from.UserID != 0 && to.ContactID == 0 && to.UserID == 0")
	}
	if input.To.UserID != "" && input.From.ContactID == "" && input.From.UserID == "" {
		return errors.New("to.UserID != 0 && from.ContactID == 0 && from.UserID == 0")
	}

	if input.From.UserID == input.To.UserID {
		if input.From.UserID == "" && input.To.UserID == "" {
			if input.From.ContactID == "" {
				return errors.New("from.UserID == 0 && to.UserID == 0 && from.ContactID == 0")
			}
			if input.To.ContactID == "" {
				return errors.New("from.UserID == 0 && to.UserID == 0 && to.ContactID == 0")
			}
		} else {
			return errors.New("from.UserID == to.UserID")
		}
	}
	switch input.CreatorUser.ID {
	case input.From.UserID:
		if input.To.ContactID == "" {
			return errors.New("creatorUserID == from.UserID && to.ContactID == 0")
		}
	case input.To.UserID:
		if input.From.ContactID == "" {
			return errors.New("creatorUserID == from.UserID && from.ContactID == 0")
		}
	default:
		if input.From.ContactID == "" {
			return errors.New("3d party transfer and from.ContactID == 0")
		}
		if input.To.ContactID == "" {
			return errors.New("3d party transfer and to.ContactID == 0")
		}
	}
	return nil
}

func (input CreateTransferInput) String() string {
	return fmt.Sprintf("CreatorUserID=%s, IsReturn=%v, ReturnToTransferID=%s, Amount=%v, From=%v, To=%v, DueOn=%v",
		input.CreatorUser.ID, input.Request.IsReturn, input.Request.ReturnToTransferID, input.Request.Amount, input.From, input.To, input.Request.DueOn)
}

func NewTransferInput(
	env string,
	source dtdal.TransferSource,
	creatorUser dbo4userus.UserEntry,
	request CreateTransferRequest,
	from, to *models4debtus.TransferCounterpartyInfo,
) (input CreateTransferInput) {
	// All checks are in the input.Validate()
	input = CreateTransferInput{
		Env:         env,
		Source:      source,
		CreatorUser: creatorUser,
		Request:     request,
		From:        from,
		To:          to,
	}
	if err := input.Validate(); err != nil {
		panic(err)
	}
	return
}
