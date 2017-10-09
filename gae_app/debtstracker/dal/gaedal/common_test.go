package gaedal

import (
	"google.golang.org/appengine/datastore"
	"testing"
)

func testDatastoreIntKey(t *testing.T, expectedID int64, key *datastore.Key) {
	if key == nil {
		t.Error("key is nil")
		return
	}
	if key.StringID() != "" {
		t.Error("StringID() is not empty")
	}
	if key.IntID() != expectedID {
		t.Error("IntegerID() != expectedID", expectedID)
	}
	if key.Parent() != nil {
		t.Error("Parent() != nil")
	}
}

func testDatastoreStringKey(t *testing.T, expectedID string, key *datastore.Key) {
	if key == nil {
		t.Error("key is nil")
		return
	}
	if key.StringID() != expectedID {
		t.Error("StringID() != expectedID", key.StringID(), expectedID)
	}
	if key.IntID() != 0 {
		t.Error("IntegerID() != 0")
	}
	if key.Parent() != nil {
		t.Error("Parent() != nil")
	}
}

func testDatastoreIncompleteKey(t *testing.T, key *datastore.Key) {
	if key == nil {
		t.Error("key is nil")
		return
	}
	if key.StringID() != "" {
		t.Error("StringID() is not empty")
	}
	if key.IntID() != 0 {
		t.Error("IntegerID() != 0")
	}
	if key.Parent() != nil {
		t.Error("Parent() != nil")
	}
}
