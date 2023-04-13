package gaedal

import (
	"testing"

	"context"
)

func TestNewLoginPinKey(t *testing.T) {
	const loginPinID = 157
	testIntKey(t, loginPinID, NewLoginPinKey(context.Background(), loginPinID))
}

func TestNewLoginPinIncompleteKey(t *testing.T) {
	testIncompleteKey(t, NewLoginPinIncompleteKey(context.Background()))
}
