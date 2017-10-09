package gaedal

import (
	"golang.org/x/net/context"
	"testing"
)

func TestNewGoogleUserKey(t *testing.T) {
	const googleUserID = "246"
	testDatastoreStringKey(t, googleUserID, NewUserGoogleKey(context.Background(), googleUserID))
}
