package gaedal

import (
	"testing"

	"golang.org/x/net/context"
)

func TestNewAppUserKey(t *testing.T) {
	const appUserID = 1234
	testDatastoreIntKey(t, appUserID, NewAppUserKey(context.Background(), appUserID))
}
