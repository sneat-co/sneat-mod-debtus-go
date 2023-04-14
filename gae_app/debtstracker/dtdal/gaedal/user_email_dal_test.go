package gaedal

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
	"testing"
)

func TestNewUserEmailKey(t *testing.T) {
	const email = "test@example.come"
	testDatastoreStringKey(t, email, models.NewUserEmailKey(email))
}

func TestUserEmailGaeDal_GetUserEmailByID(t *testing.T) {
	//gaedb.Get = func(c context.Context, key *dal.Key, val interface{}) error {
	//	return nil
	//}

	userEmail, _ := NewUserEmailGaeDal().GetUserEmailByID(context.Background(), nil, " JackSmith@Example.com ")

	if userEmail.ID != "jacksmith@example.com" {
		t.Error("userEmail.ID expected to be lower case without spaces")
	}
}
