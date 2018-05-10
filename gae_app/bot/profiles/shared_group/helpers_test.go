package shared_group

import "testing"

func TestGetGroup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should panic")
		}
	}()
	GetGroup(nil, nil)
}
