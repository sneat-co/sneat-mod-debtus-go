package gaedal

import (
	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/models"
	"testing"
)

func TestNewReceiptIncompleteKey(t *testing.T) {
	testIncompleteKey(t, models.NewReceiptIncompleteKey())
}

func TestNewReceiptKey(t *testing.T) {
	const receiptID = 234
	testIntKey(t, receiptID, models.NewReceiptKey(receiptID))
}
