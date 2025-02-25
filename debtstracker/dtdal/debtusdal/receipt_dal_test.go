package debtusdal

import (
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"testing"
)

func TestNewReceiptIncompleteKey(t *testing.T) {
	testIncompleteKey(t, models4debtus.NewReceiptIncompleteKey())
}

func TestNewReceiptKey(t *testing.T) {
	const receiptID = "234"
	testStrKey(t, receiptID, models4debtus.NewReceiptKey(receiptID))
}
