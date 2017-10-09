package gaestandard

import (
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
)

var defaultVersionHostname = appengine.DefaultVersionHostname

func GetEnvironment(c context.Context) strongo.Environment {
	hostname := defaultVersionHostname(c)
	return common.GetEnvironmentFromHost(hostname)
}
