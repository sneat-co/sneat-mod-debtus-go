package bot

import (
	"testing"
	"github.com/strongo/bots-framework/core"
)

func TestSendRefreshOrNothingChanged(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()

	var m bots.MessageFromBot
	SendRefreshOrNothingChanged(nil, m)
}
