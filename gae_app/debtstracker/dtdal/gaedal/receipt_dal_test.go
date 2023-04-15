package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"testing"
)

func TestNewReceiptIncompleteKey(t *testing.T) {
	testIncompleteKey(t, models.NewReceiptIncompleteKey())
}

func TestNewReceiptKey(t *testing.T) {
	const receiptID = 234
	testIntKey(t, receiptID, models.NewReceiptKey(receiptID))
}
