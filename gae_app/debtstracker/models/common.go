package models

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/appengine/datastore"
)

const (
	STATUS_ACTIVE   = "active"
	STATUS_DRAFT    = "draft"
	STATUS_DELETED  = "deleted"
	STATUS_ARCHIVED = "archived"
)

func validateString(errMess, s string, validValues []string) error {
	var ok bool
	for _, validValue := range validValues {
		if s == validValue {
			ok = true
		}
	}
	if !ok {
		return fmt.Errorf("%v: '%v'", errMess, s)
	}
	return nil
}

var ErrNoProperties = errors.New("No properties")

var checkHasProperties = func(kind string, properties []datastore.Property) error {
	if len(properties) == 0 {
		panic(ErrNoProperties.Error())
	}
	return nil
}
