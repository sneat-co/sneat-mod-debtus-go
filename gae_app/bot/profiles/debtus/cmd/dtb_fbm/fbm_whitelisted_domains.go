package dtb_fbm

import (
	"github.com/strongo/bots-api-fbm"
	"github.com/strongo/app"
	"fmt"
	"github.com/strongo/bots-framework/core"
	"golang.org/x/net/context"
	"net/http"
)

func SetWhitelistedDomains(c context.Context, r *http.Request, bot bots.BotSettings, api fbm_api.GraphAPI) (err error) {
	var whitelistedDomainsMessage fbm_api.WhitelistedDomainsMessage
	switch bot.Env {
	case strongo.EnvProduction:
		whitelistedDomainsMessage = fbm_api.WhitelistedDomainsMessage{WhitelistedDomains: []string{
			"https://debtstracker.io",
			"https://splitbill.co",
		}}
	case strongo.EnvLocal:
		domains := []string{
			"https://debtstracker.local",
		}
		host := r.URL.Query().Get("host")
		if host != "" {
			domains = append(domains, fmt.Sprintf("https://%v", host))
		}
		whitelistedDomainsMessage = fbm_api.WhitelistedDomainsMessage{WhitelistedDomains: domains}
	case strongo.EnvDevTest:
		whitelistedDomainsMessage = fbm_api.WhitelistedDomainsMessage{WhitelistedDomains: []string{
			"https://debtstracker-dev1.appspot.com",
		}}
	default:
		err = fmt.Errorf("Unknown bot environment: %d=%v", bot.Env, strongo.EnvironmentNames[bot.Env])
		return
	}

	host := fmt.Sprintf("https://%v", r.Host)

	hasHost := false
	for _, v := range whitelistedDomainsMessage.WhitelistedDomains {
		if v == host {
			hasHost = true
			break
		}
	}
	if !hasHost {
		whitelistedDomainsMessage.WhitelistedDomains = append(whitelistedDomainsMessage.WhitelistedDomains, host)
	}

	if err = api.SetWhitelistedDomains(c, whitelistedDomainsMessage); err != nil {
		return
	}
	return
}
