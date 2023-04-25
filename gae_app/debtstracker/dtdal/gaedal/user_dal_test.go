package gaedal

import (
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"testing"
)

func TestNewAppUserKey(t *testing.T) {
	const appUserID = 1234
	testIntKey(t, appUserID, models.NewAppUserKey(appUserID))
}
