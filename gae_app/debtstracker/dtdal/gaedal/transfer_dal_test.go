package gaedal

import (
	"testing"

	"context"
)

func TestNewTransferKey(t *testing.T) {
	const transferID = 12345
	testIntKey(t, transferID, NewTransferKey(context.Background(), transferID))
}
