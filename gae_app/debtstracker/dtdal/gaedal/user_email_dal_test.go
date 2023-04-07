package gaedal

import (
	"testing"

	"context"
	"github.com/strongo/db/gaedb"
	"google.golang.org/appengine/datastore"
)

func TestNewUserEmailKey(t *testing.T) {
	const email = "test@example.come"
	testDatastoreStringKey(t, email, NewUserEmailKey(context.Background(), email))
}

func TestUserEmailGaeDal_GetUserEmailByID(t *testing.T) {
	gaedb.Get = func(c context.Context, key *datastore.Key, val interface{}) error {
		return nil
	}

	userEmail, _ := NewUserEmailGaeDal().GetUserEmailByID(context.Background(), " JackSmith@Example.com ")

	if userEmail.ID != "jacksmith@example.com" {
		t.Error("userEmail.ID expected to be lower case without spaces")
	}
}
