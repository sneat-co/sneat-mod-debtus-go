package gaedal

import (
	"testing"

	"context"
)

func TestNewLoginPinKey(t *testing.T) {
	const loginPinID = 157
	testDatastoreIntKey(t, loginPinID, NewLoginPinKey(context.Background(), loginPinID))
}

func TestNewLoginPinIncompleteKey(t *testing.T) {
	testDatastoreIncompleteKey(t, NewLoginPinIncompleteKey(context.Background()))
}
