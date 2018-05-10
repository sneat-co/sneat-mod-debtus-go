package dto

import (
	"testing"
	"github.com/pquerna/ffjson/ffjson"
)

func TestContactDto_MarshalJSON(t *testing.T) {
	contact := ContactDto{}
	_, err := ffjson.MarshalFast(&contact)
	if err != nil {
		t.Fatal(err)
	}
}