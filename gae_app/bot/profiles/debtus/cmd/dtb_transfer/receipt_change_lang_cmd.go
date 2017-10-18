package dtb_transfer

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/pkg/errors"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
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


