package gaedal

import (
	"testing"

	"context"
)

func TestNewReceiptIncompleteKey(t *testing.T) {
	testDatastoreIncompleteKey(t, NewReceiptIncompleteKey(context.Background()))
}

func TestNewReceiptKey(t *testing.T) {
	const receiptID = 234
	testDatastoreIntKey(t, receiptID, NewReceiptKey(context.Background(), receiptID))
}
