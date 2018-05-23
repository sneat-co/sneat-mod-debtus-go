package dtb_transfer

import (
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
)

const CHANGE_RECEIPT_LANG_COMMAND = "change-lang-receipt"

var ChangeReceiptAnnouncementLangCommand = bots.NewCallbackCommand(
	CHANGE_RECEIPT_LANG_COMMAND,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		query := callbackUrl.Query()
		code5 := query.Get("locale")
		if len(code5) != 5 {
			return m, errors.New("ChangeReceiptAnnouncementLangCommand: len(code5) != 5")
		}
		whc.SetLocale(code5)
		receiptID, err := common.DecodeID(query.Get("id"))
		if err != nil {
			return m, err
		}
		return showReceiptAnnouncement(whc, receiptID, "")
	},
)
