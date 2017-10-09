package models

import (
	"github.com/matryer/is"
	"testing"
	"time"
)

func TestAppUserEntity_Contacts(t *testing.T) {
	var userEntity AppUserEntity

	userEntity.ContactsJsonActive = `[{"ID":1,"Name":"Alex (Alex)"}]`

	contacts := userEntity.Contacts()

	contact := contacts[0]
	is := is.New(t)
	is.Equal(contact.Name, "Alex")
	is.Equal(contact.Status, "active")
}

func TestAppUserEntity_SetLastCurrency(t *testing.T) {
	userEntity := AppUserEntity{}
	userEntity.SetLastCurrency("EUR")
	if len(userEntity.LastCurrencies) != 1 {
		t.Errorf("Expected 1 value in LastCurrencies, got: %d", len(userEntity.LastCurrencies))
	}
	userEntity.SetLastCurrency("USD")
	if len(userEntity.LastCurrencies) != 2 {
		t.Errorf("Expected 2 values in LastCurrencies, got: %d", len(userEntity.LastCurrencies))
	}
	if userEntity.LastCurrencies[0] != "USD" {
		t.Errorf("First currency should be USD, got: %v", userEntity.LastCurrencies[0])
	}
	if userEntity.LastCurrencies[1] != "EUR" {
		t.Errorf("Second currency should be EUR, got: %v", userEntity.LastCurrencies[1])
	}

	userEntity.SetLastCurrency("EUR")
	if len(userEntity.LastCurrencies) != 2 {
		t.Errorf("Expected 2 values in LastCurrencies, got: %d", len(userEntity.LastCurrencies))
	}
	if userEntity.LastCurrencies[0] != "EUR" {
		t.Errorf("Second currency should be EUR, got: %v", userEntity.LastCurrencies[0])
	}
	if userEntity.LastCurrencies[1] != "USD" {
		t.Errorf("First currency should be USD, got: %v", userEntity.LastCurrencies[1])
	}
}

func TestLastLogin_SetLastLogin(t *testing.T) {
	user := NewUser(ClientInfo{})
	now := time.Now()
	user.SetLastLogin(now)
	if user.DtLastLogin != now {
		t.Errorf("user.DtLastLogin != now")
	}

	userGoogle := UserGoogle{
		UserGoogleEntity: &UserGoogleEntity{},
	}
	userGoogle.SetLastLogin(now)
	if userGoogle.DtLastLogin != now {
		t.Errorf("userGoogle.DtLastLogin != now")
	}

	type LastLoginSetter interface {
		SetLastLogin(v time.Time)
	}

	userGoogle = UserGoogle{
		UserGoogleEntity: &UserGoogleEntity{},
	}
	var lastLoginSetter LastLoginSetter = userGoogle
	lastLoginSetter.SetLastLogin(now)
	if userGoogle.DtLastLogin != now {
		t.Errorf("lastLoginSetter.DtLastLogin != now")
	}
}
