package api

//go:generate ffjson $GOFILE

import "bitbucket.com/debtstracker/gae_app/debtstracker/models"

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
