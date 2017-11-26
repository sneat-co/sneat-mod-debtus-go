package gaedal

import (
	"testing"

	"golang.org/x/net/context"
)

func TestNewGroupKey(t *testing.T) {
	const groupID = 456
	testDatastoreIntKey(t, groupID, NewGroupKey(context.Background(), groupID))
}
