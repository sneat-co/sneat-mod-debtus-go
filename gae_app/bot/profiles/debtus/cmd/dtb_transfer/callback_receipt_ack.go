package dtb_transfer

import (
	"github.com/bots-go-framework/bots-fw/botsfw"
	"net/url"

	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
)

const ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND = "ack-receipt"

var AcknowledgeReceiptCallbackCommand = botsfw.NewCallbackCommand(ACKNOWLEDGE_RECEIPT_CALLBACK_COMMAND, func(whc botsfw.WebhookContext, callbackUrl *url.URL) (m botsfw.MessageFromBot, err error) {
	query := callbackUrl.Query()
	receiptID, err := common.DecodeIntID(query.Get("id"))
	if err != nil {
		return m, err
	}

	return AcknowledgeReceipt(whc, receiptID, query.Get("do"))
})
