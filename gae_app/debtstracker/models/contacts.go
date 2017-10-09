package models

import (
	"fmt"
	"strconv"
	"strings"
)

type PhoneContact struct {
	// Part of ContactDetails => Part of User|Contact
	// Contact details
	PhoneNumber            int64
	PhoneNumberConfirmed   bool
	PhoneNumberIsConfirmed bool `datastore:",noindex"` // Deprecated
	//+9223372036854775807
	//+353857403948
	//+79169743259
}

func (p PhoneContact) PhoneNumberAsString() string {
	return "+" + strconv.FormatInt(p.PhoneNumber, 10)
}

type EmailContact struct {
	EmailAddress         string
	EmailAddressOriginal string `datastore:",noindex"`
	EmailConfirmed       bool   `datastore:",noindex"`
}

func (ec *EmailContact) SetEmail(email string, confirmed bool) EmailContact {
	ec.EmailAddress = strings.ToLower(email)
	if ec.EmailAddress != email {
		ec.EmailAddressOriginal = email
	} else {
		ec.EmailAddressOriginal = ""
	}
	ec.EmailConfirmed = confirmed
	return *ec
}

type ContactDetails struct {
	// Helper struct, not stored as independent entity
	PhoneContact
	EmailContact
	FirstName      string `datastore:",noindex"`
	LastName       string `datastore:",noindex"`
	ScreenName     string `datastore:",noindex"`
	Nickname       string `datastore:",noindex"`
	Username       string `datastore:",noindex"` //TODO: Should it be "Name"?
	TelegramUserID int64  // When user ads Telegram contact we store Telegram user_id so we can link users later.
}

func (contact *ContactDetails) FullName() string {
	if contact.LastName != "" && contact.FirstName != "" {
		if contact.Username == "" || strings.ToLower(contact.FirstName) == strings.ToLower(contact.Username) || strings.ToLower(contact.LastName) == strings.ToLower(contact.Username) {
			return fmt.Sprintf("%v %v", contact.FirstName, contact.LastName)
		} else {
			return fmt.Sprintf("%v %v (%v)", contact.FirstName, contact.LastName, contact.Username)
		}

	} else if contact.FirstName != "" {
		if contact.Username == "" || contact.Username == contact.FirstName {
			return contact.FirstName
		} else {
			return fmt.Sprintf("%v (%v)", contact.FirstName, contact.Username)
		}
	} else if contact.LastName != "" {
		if contact.Username == "" || contact.Username == contact.LastName {
			return contact.FirstName
		} else {
			return fmt.Sprintf("%v (%v)", contact.LastName, contact.Username)
		}
	} else if contact.ScreenName != "" {
		return contact.ScreenName
	} else if contact.Username != "" {
		return contact.Username
	} else {
		return NO_NAME
	}
}

const NO_NAME = ">NO_NAME<"
