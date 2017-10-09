package gae_app

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"testing"
)

func TestInitBot(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should fail")
		}
	}()
	InitBots(nil, nil, common.DebtsTrackerAppContext{})
}
