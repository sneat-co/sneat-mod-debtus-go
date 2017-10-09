package analytics

import (
	//"github.com/strongo/measurement-protocol"
	"testing"
)

func TestSendSingleMessage(t *testing.T) {
	if err := SendSingleMessage(nil, nil); err == nil {
		t.Error("Expected to get error on nil context")
	}
}
