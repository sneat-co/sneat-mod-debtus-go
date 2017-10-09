package dtb_general

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"bitbucket.com/debtstracker/gae_app/debtstracker/models"
	"fmt"
	"github.com/DebtsTracker/translations/trans"
	"github.com/strongo/bots-framework/core"
)

func EditReminderMessage(whc bots.WebhookContext, transfer models.Transfer, message string) (m bots.MessageFromBot, err error) {
	utm := common.NewUtmParams(whc, common.UTM_CAMPAIGN_REMINDER)
	mt := fmt.Sprintf(
		"<b>%v</b>\n%v\n\n%v",
		whc.Translate(trans.MESSAGE_TEXT_REMINDER),
		common.TextReceiptForTransfer(whc, transfer, whc.AppUserIntID(), common.ShowReceiptToAutodetect, utm),
		message,
	)
	if whc.InputType() == bots.WebhookInputCallbackQuery {
		if m, err = whc.NewEditMessage(mt, bots.MessageFormatHTML); err != nil {
			return
		}
	} else {
		m = whc.NewMessage(mt)
		m.Format = bots.MessageFormatHTML
		SetMainMenuKeyboard(whc, &m)
	}

	return
}

