package common

import (
	isLib "github.com/matryer/is"
	"testing"
)

func TestDecodeID(t *testing.T) {
	is := isLib.New(t)

	_, err := DecodeID("")
	is.True(err != nil) // Should return error if empty string
}
