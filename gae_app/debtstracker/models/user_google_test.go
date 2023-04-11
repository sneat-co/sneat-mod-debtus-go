package models

import (
	"testing"

	isLib "github.com/matryer/is"
)

func TestUserGoogleEntity_GetEmail(t *testing.T) {
	is := isLib.New(t)

	entity := UserGoogleData{}
	entity.Email = "test@example.com"
	is.Equal(entity.Email, entity.GetEmail())
}
