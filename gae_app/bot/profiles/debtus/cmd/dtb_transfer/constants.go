package dtb_transfer

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/common"
	"fmt"
)

const (
	RECEIPT_ACTION__DO_NOT_SEND    = "do-not-send"
	SEND_RECEIPT_CALLBACK_PATH     = "send-receipt"
	SEND_RECEIPT_BY_CHOOSE_CHANNEL = "select"
	WIZARD_PARAM_TRANSFER          = "transfer"
	WIZARD_PARAM_REMINDER          = "reminder"
	WIZARD_PARAM_COUNTERPARTY      = "counterparty" // TODO: Decide use this or WIZARD_PARAM_CONTACT
	WIZARD_PARAM_CONTACT           = "contact"      // TODO: Decide use this or WIZARD_PARAM_COUNTERPARTY
)

type SendReceipt struct {
	TransferID int64
	By         string
}

func SendReceiptCallbackData(transferID int64, by string) string {
	return fmt.Sprintf("%v?by=%v&transfer=%v", SEND_RECEIPT_CALLBACK_PATH, by, common.EncodeID(transferID))
}

func SendReceiptUrl(transferID int64, by string) string {
	return fmt.Sprintf("https://debtstracker.io/app/send-receipt?by=%v&transfer=%v", by, common.EncodeID(transferID))
}
