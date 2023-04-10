package bot

import (
	"testing"
)

func TestSendRefreshOrNothingChanged(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()

	var m botsfw.MessageFromBot
	SendRefreshOrNothingChanged(nil, m)
}
