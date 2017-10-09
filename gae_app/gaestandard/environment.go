package gaestandard

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/app"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

var defaultVersionHostname = appengine.DefaultVersionHostname

func GetEnvironment(c context.Context) strongo.Environment {
	hostname := defaultVersionHostname(c)
	return common.GetEnvironmentFromHost(hostname)
}
