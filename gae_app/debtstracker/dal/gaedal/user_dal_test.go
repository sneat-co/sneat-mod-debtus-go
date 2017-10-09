package gaedal

import (
	"golang.org/x/net/context"
	"testing"
)

func TestNewAppUserKey(t *testing.T) {
	const appUserID = 1234
	testDatastoreIntKey(t, appUserID, NewAppUserKey(context.Background(), appUserID))
}
