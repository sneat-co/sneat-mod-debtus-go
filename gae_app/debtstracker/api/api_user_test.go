package api

import (
	"testing"

	"golang.org/x/net/context"
)

func TestApiUserInfo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()

	c := context.Background()
	handleUserInfo(c, nil, nil)
}
