package common

import (
	"github.com/strongo/app"
	"strings"
)

func GetEnvironmentFromHost(host string) strongo.Environment {
	if host == "debtstracker.io" || host == "debtstracker-io.appspot.com" {
		return strongo.EnvProduction
	} else if strings.HasPrefix(host, "debtstracker-dev") && strings.HasSuffix(host, ".appspot.com") {
		return strongo.EnvDevTest
	} else if host == "localhost" || strings.HasPrefix(host,"localhost:") || strings.HasSuffix(host,".ngrok.io") || host == "debtstracker.local" {
		return strongo.EnvLocal
	}
	return strongo.EnvUnknown
}
