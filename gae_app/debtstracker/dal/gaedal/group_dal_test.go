package gaedal

import (
	"golang.org/x/net/context"
	"testing"
)

func TestNewGroupKey(t *testing.T) {
	const groupID = 456
	testDatastoreIntKey(t, groupID, NewGroupKey(context.Background(), groupID))
}
