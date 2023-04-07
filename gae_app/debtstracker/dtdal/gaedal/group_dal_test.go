package gaedal

import (
	"testing"

	"context"
)

func TestNewGroupKey(t *testing.T) {
	const groupID = "456"
	testDatastoreStringKey(t, groupID, NewGroupKey(context.Background(), groupID))
}
