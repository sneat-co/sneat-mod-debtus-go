package dtb_transfer

import (
	"net/url"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/common"
	"github.com/strongo/bots-framework/core"
)

const ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND = "ack-receipt"

var AcknowledgeReceiptCallbackCommand = bots.NewCallbackCommand(ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND, func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
	query := callbackUrl.Query()
	receiptID, err := common.DecodeID(query.Get("id"))
	if err != nil {
		return m, err
	}

	return AcknowledgeReceipt(whc, receiptID, query.Get("do"))
})
