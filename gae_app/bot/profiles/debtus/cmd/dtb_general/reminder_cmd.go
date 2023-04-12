package dtb_general

import (
	"fmt"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/sneat-co/debtstracker-translations/trans"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
)

func EditReminderMessage(whc botsfw.WebhookContext, transfer models.Transfer, message string) (m botsfw.MessageFromBot, err error) {
	utm := common.NewUtmParams(whc, common.UTM_CAMPAIGN_REMINDER)
	mt := fmt.Sprintf(
		"<b>%v</b>\n%v\n\n%v",
		whc.Translate(trans.MESSAGE_TEXT_REMINDER),
		common.TextReceiptForTransfer(whc, transfer, whc.AppUserIntID(), common.ShowReceiptToAutodetect, utm),
		message,
	)
	if whc.InputType() == botsfw.WebhookInputCallbackQuery {
		if m, err = whc.NewEditMessage(mt, botsfw.MessageFormatHTML); err != nil {
			return
		}
	} else {
		m = whc.NewMessage(mt)
		m.Format = botsfw.MessageFormatHTML
		SetMainMenuKeyboard(whc, &m)
	}

	return
}
