package gaeapp

import (
	"testing"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
)

func TestInitBot(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should fail")
		}
	}()
	InitBots(nil, nil, common.DebtsTrackerAppContext{})
}
