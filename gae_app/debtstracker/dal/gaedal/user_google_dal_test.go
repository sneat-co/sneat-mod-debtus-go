package gaedal

import (
	"testing"

	"golang.org/x/net/context"
)

func TestNewGoogleUserKey(t *testing.T) {
	const googleUserID = "246"
	testDatastoreStringKey(t, googleUserID, NewUserGoogleKey(context.Background(), googleUserID))
}
