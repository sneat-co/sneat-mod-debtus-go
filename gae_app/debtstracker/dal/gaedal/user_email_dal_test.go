package gaedal

import (
	"testing"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"github.com/strongo/app/gaedb"
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