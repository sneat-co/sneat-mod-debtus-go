package dtb_general

import (
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"fmt"
	"github.com/sneat-co/debtstracker-translations/trans"
	"strings"
)

func AdSlot(whc botsfw.WebhookContext, place string) string {
	utmParams := common.FillUtmParams(whc, common.UtmParams{Campaign: place})
	link := fmt.Sprintf(`href="https://debtstracker.io/%v/ads#%v"`, whc.Locale().SiteCode(), utmParams)
	return strings.Replace(whc.Translate(trans.MESSAGE_TEXT_YOUR_AD_COULD_BE_HERE), "href", link, 1)
}
