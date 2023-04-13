package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"testing"
)

func TestNewAppUserKey(t *testing.T) {
	const appUserID = 1234
	testIntKey(t, appUserID, models.NewAppUserKey(appUserID))
}
