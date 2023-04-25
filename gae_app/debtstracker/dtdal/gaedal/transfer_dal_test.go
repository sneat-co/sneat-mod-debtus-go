package gaedal

import (
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"testing"
)

func TestNewTransferKey(t *testing.T) {
	const transferID = 12345
	testIntKey(t, transferID, models.NewTransferKey(transferID))
}
