package appengine

import (
	"github.com/strongo/app/log"
	"testing"
)

func TestInit(t *testing.T) {
	if log.NumberOfLoggers() == 0 {
		t.Error("At least 1 logger should be added")
	}
}
