package vkbots

import "testing"

func TestBotsBy(t *testing.T) {
	if len(BotsBy.ByCode) == 0 {
		t.Fatal("len(BotsBy.ByCode) == 0")
	}
}
