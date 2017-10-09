package models

import (
	isLib "github.com/matryer/is"
	"testing"
)

func TestUserGoogleEntity_GetEmail(t *testing.T) {
	is := isLib.New(t)

	entity := UserGoogleEntity{}
	entity.Email = "test@example.com"
	is.Equal(entity.Email, entity.GetEmail())
}
