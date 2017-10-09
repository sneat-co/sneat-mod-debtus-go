package gaedal

import (
	"golang.org/x/net/context"
	"testing"
)

func TestNewTransferKey(t *testing.T) {
	const transferID = 12345
	testDatastoreIntKey(t, transferID, NewTransferKey(context.Background(), transferID))
}
