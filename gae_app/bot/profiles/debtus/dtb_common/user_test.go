package dtb_common

import "testing"

func TestGetUserWithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()
	GetUser(nil)
}
