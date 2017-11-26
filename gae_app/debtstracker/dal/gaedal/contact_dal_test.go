package gaedal

import (
	"testing"

	isLib "github.com/matryer/is"
	"golang.org/x/net/context"
)

func TestNewContactIncompleteKey(t *testing.T) {
	is := isLib.New(t)
	c := context.Background()
	key := NewContactIncompleteKey(c)
	is.True(key.IntID() == 0)
	is.True(key.StringID() == "")
	is.True(key.Parent() == nil)
}

func TestNewContactKey(t *testing.T) {
	is := isLib.New(t)
	c := context.Background()
	key := NewContactKey(c, 135)
	is.True(key.IntID() == 135)
	is.True(key.StringID() == "")
	is.True(key.Parent() == nil)
}
