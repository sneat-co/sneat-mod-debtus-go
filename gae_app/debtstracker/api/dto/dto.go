package dto

//go:generate ffjson $GOFILE

import (
	"time"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/strongo/decimal"
	"encoding/json"
	"github.com/pquerna/ffjson/ffjson"
)

type UserMeDto struct {
	UserID       int64
	FullName     string `json:",omitempty"`
	GoogleUserID string `json:",omitempty"`
	FbUserID     string `json:",omitempty"`
	VkUserID     int64  `json:",omitempty"`
	ViberUserID  string `json:",omitempty"`
}



type ApiAcknowledgeDto struct {
	Status   string
	UnixTime int64
}

type ApiReceiptDto struct {
	ID       int64 `json:"Id"`
	Code     string
	Transfer ApiReceiptTransferDto
	SentVia  string
	SentTo   string `json:",omitempty"`
}

type ApiUserDto struct {
	ID   int64  `json:"Id"`
	Name string `json:",omitempty"`
}

type ApiReceiptTransferDto struct {
	// TODO: We are not replacing with TransferDto as it has From/To => Creator optimisation. Think if we can reuse.
	ID             int64 `json:"Id"`
	Amount         models.Amount
	From           ContactDto
	DtCreated      time.Time
	To             ContactDto
	IsOutstanding  bool
	Creator        ApiUserDto
	CreatorComment string             `json:",omitempty"`
	Acknowledge    *ApiAcknowledgeDto `json:",omitempty"`
}


type ContactDto struct {
	ID     int64 `json:",omitempty"` // TODO: Document why it can be empty?
	UserID int64 `json:",omitempty"`
	Name   string `json:",omitempty"`
	//Note string `json:",omitempty"`
	Comment string `json:",omitempty"`
}

func NewContactDto(counterpartyInfo models.TransferCounterpartyInfo) ContactDto {
	dto := ContactDto{
		ID:      counterpartyInfo.ContactID,
		UserID:  counterpartyInfo.UserID,
		Name:    counterpartyInfo.Name(),
		Comment: counterpartyInfo.Comment,
	}
	if dto.Name == models.NO_NAME {
		dto.Name = ""
	}
	return dto
}

type BillDto struct {
	// TODO: Generate ffjson
	ID      string
	Name    string
	Amount  models.Amount
	Members []BillMemberDto
}

type BillMemberDto struct {
	UserID     int64               `json:",omitempty"`
	ContactID  int64               `json:",omitempty"`
	Amount     decimal.Decimal64p2
	Paid       decimal.Decimal64p2 `json:",omitempty"`
	Share      int                 `json:",omitempty"`
	Adjustment decimal.Decimal64p2 `json:",omitempty"`
}

type ContactListDto struct {
	ContactDto
	Status  string
	Balance *json.RawMessage `json:",omitempty"`
}

type EmailInfo struct {
	Address     string
	IsConfirmed bool
}

type PhoneInfo struct {
	Number      int64
	IsConfirmed bool
}

type ContactDetailsDto struct {
	ContactListDto
	Email  *EmailInfo `json:",omitempty"`
	Phone  *PhoneInfo `json:",omitempty"`
	TransfersResultDto
	Groups []ContactGroupDto `json:",omitempty"`
}

type TransfersResultDto struct {
	HasMoreTransfers bool `json:",omitempty"`
	Transfers        []*TransferDto `json:",omitempty"`
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

func TransfersToDto(userID int64, transfers []models.Transfer) []*TransferDto {
	transfersDto := make([]*TransferDto, len(transfers))
	for i, transfer := range transfers {
		transfersDto[i] = TransferToDto(userID, transfer)
	}
	return transfersDto
}

type CreateTransferResponse struct {
	Error               string           `json:",omitempty"`
	Transfer            *TransferDto     `json:",omitempty"`
	CounterpartyBalance *json.RawMessage `json:",omitempty"`
	UserBalance         *json.RawMessage `json:",omitempty"`
}

func TransferToDto(userID int64, transfer models.Transfer) *TransferDto {
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

type GroupDto struct {
	ID           string
	Name         string
	Status       string
	Note         string           `json:",omitempty"`
	MembersCount int              `json:",omitempty"`
	Members      []GroupMemberDto `json:",omitempty"`
}


type GroupMemberDto struct {
	ID        string
	UserID    int64  `json:",omitempty"`
	ContactID int64  `json:",omitempty"`
	Name      string `json:",omitempty"`
}


type ContactGroupDto struct {
	ID           string
	Name         string
	MemberID     int64
	MembersCount int
}

type CounterpartyDto struct {
	Id      int64
	UserID  int64 `json:",omitempty"`
	Name    string
	Balance *json.RawMessage `json:",omitempty"`
}
type Record struct {
	Id                     int64
	Name                   string
	Counterparties         []CounterpartyDto
	Transfers              int
	CountOfReceiptsCreated int
	InvitedByUser          *struct {
		Id   int64
		Name string
	} `json:",omitempty"`
	//InvitedByUserID int64 `json:",omitempty"`
	//InvitedByUserName string `json:",omitempty"`
	Balance         *json.RawMessage `json:",omitempty"`
	TelegramUserIDs []int64          `json:",omitempty"`
}
