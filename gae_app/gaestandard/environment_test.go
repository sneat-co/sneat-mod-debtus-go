package gaestandard

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"testing"
)

func TestGetEnvironment(t *testing.T) {

	testEnv := func(host string, expected strongo.Environment) {
		defaultVersionHostname = func(c context.Context) string {
			return host
		}
		if environment := GetEnvironment(context.Background()); environment != expected {
			t.Errorf("Unexpected environment: %v", environment)
		}
	}

	testEnv("debtstracker-io.appspot.com", strongo.EnvProduction)
	testEnv("debtstracker.local", strongo.EnvLocal)
}
