package dtmocks

import (
	"testing"
	"context"
)

func TestSetupMocks(t *testing.T) {
	c := context.Background()
	SetupMocks(c)
}
