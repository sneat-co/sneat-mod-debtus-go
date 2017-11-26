package gaedal

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"
)

func TestNewUserFacebookKey(t *testing.T) {
	const (
		fbAppID  = "12345"
		fbUserID = "975"
	)
	key := NewUserFacebookKey(context.Background(), fbAppID, fbUserID)
	testDatastoreStringKey(t, fmt.Sprintf("%v:%v", fbAppID, fbUserID), key)
}
