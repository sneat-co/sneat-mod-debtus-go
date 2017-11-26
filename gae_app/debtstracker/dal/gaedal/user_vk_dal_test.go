package gaedal

import (
	"testing"

	"golang.org/x/net/context"
)

func TestNewVkUserKey(t *testing.T) {
	const vkUserID = 789
	testDatastoreIntKey(t, vkUserID, NewUserVkKey(context.Background(), vkUserID))
}
