package gaedal

import (
	"testing"

	"context"
)

func TestNewAppUserKey(t *testing.T) {
	const appUserID = 1234
	testDatastoreIntKey(t, appUserID, NewAppUserKey(context.Background(), appUserID))
}
