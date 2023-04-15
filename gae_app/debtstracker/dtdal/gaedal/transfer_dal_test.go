package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"testing"
)

func TestNewTransferKey(t *testing.T) {
	const transferID = 12345
	testIntKey(t, transferID, models.NewTransferKey(transferID))
}
