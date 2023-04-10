package dtb_transfer

import (
	"net/url"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/common"
)

const ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND = "ack-receipt"

var AcknowledgeReceiptCallbackCommand = botsfw.NewCallbackCommand(ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND, func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
	query := callbackUrl.Query()
	receiptID, err := common.DecodeID(query.Get("id"))
	if err != nil {
		return m, err
	}

	return AcknowledgeReceipt(whc, receiptID, query.Get("do"))
})
