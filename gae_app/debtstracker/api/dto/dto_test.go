package dto

import (
	"github.com/pquerna/ffjson/ffjson"
	"testing"
)

func TestContactDto_MarshalJSON(t *testing.T) {
	contact := ContactDto{}
	_, err := ffjson.MarshalFast(&contact)
	if err != nil {
		t.Fatal(err)
	}
}
