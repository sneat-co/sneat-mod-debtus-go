package gaedal

import (
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/models"
	"testing"
)

func TestNewTransferKey(t *testing.T) {
	const transferID = "12345"
	testStrKey(t, transferID, models.NewTransferKey(transferID))
}
