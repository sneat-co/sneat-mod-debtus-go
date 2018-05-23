package dtb_general

import (
	"fmt"
	"strings"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
)

func AdSlot(whc bots.WebhookContext, place string) string {
	utmParams := common.FillUtmParams(whc, common.UtmParams{Campaign: place})
	link := fmt.Sprintf(`href="https://debtstracker.io/%v/ads#%v"`, whc.Locale().SiteCode(), utmParams)
	return strings.Replace(whc.Translate(trans.MESSAGE_TEXT_YOUR_AD_COULD_BE_HERE), "href", link, 1)
}
