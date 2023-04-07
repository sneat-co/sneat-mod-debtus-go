package gaedal

import (
	"testing"

	"context"
)

func TestNewGoogleUserKey(t *testing.T) {
	const googleUserID = "246"
	testDatastoreStringKey(t, googleUserID, NewUserGoogleKey(context.Background(), googleUserID))
}
