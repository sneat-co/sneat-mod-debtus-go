package gaeapp

import (
	"testing"

	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
)

func TestInitBot(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should fail")
		}
	}()
	InitBots(nil, nil, common.DebtsTrackerAppContext{})
}
