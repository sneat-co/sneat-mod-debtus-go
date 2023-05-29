package models

import (
	"github.com/strongo/app/user"
	"testing"

	isLib "github.com/matryer/is"
)

func TestUserGoogleEntity_GetEmail(t *testing.T) {
	is := isLib.New(t)

	entity := UserAccount{
		data: &user.AccountDataBase{},
	}
	entity.data.EmailLowerCase = "test@example.com"
	is.Equal("test@example.com", entity.Data().GetEmailLowerCase())
}
