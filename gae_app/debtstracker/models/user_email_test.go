package models

import (
	"github.com/strongo/app/user"
	"testing"
)

func TestUserEmail(t *testing.T) {
	var _ user.AccountRecord = (*UserEmail)(nil)
}

func TestUserEmailEntity(t *testing.T) {
	var _ user.AccountEntity = (*UserEmailEntity)(nil)
}

func TestUserEmailEntity_AddProvider(t *testing.T) {
	entity := new(UserEmailEntity)

	if changed := entity.AddProvider("facebook"); !changed {
		t.Error("Should return changed=true")
	}
	if providerCount := len(entity.Providers); providerCount != 1 {
		t.Errorf("Expected to have 1 provider, got: %d", providerCount)
	}
	if changed := entity.AddProvider("facebook"); changed {
		t.Error("Should return changed=false")
	}
}
