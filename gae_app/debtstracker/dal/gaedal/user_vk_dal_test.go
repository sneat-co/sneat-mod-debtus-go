package gaedal

import (
	"golang.org/x/net/context"
	"testing"
)

func TestNewVkUserKey(t *testing.T) {
	const vkUserID = 789
	testDatastoreIntKey(t, vkUserID, NewUserVkKey(context.Background(), vkUserID))
}
