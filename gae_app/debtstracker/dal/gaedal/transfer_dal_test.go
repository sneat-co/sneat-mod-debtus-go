package gaedal

import (
	"testing"

	"golang.org/x/net/context"
)

func TestNewTransferKey(t *testing.T) {
	const transferID = 12345
	testDatastoreIntKey(t, transferID, NewTransferKey(context.Background(), transferID))
}
