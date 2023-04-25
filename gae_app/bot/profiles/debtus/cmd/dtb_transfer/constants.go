package dtb_transfer

import (
	"fmt"

	"github.com/sneat-co/debtstracker-go/gae_app/debtstracker/common"
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
	//transferID int
	By string
}

func SendReceiptCallbackData(transferID int, by string) string {
	return fmt.Sprintf("%v?by=%v&transfer=%v", SEND_RECEIPT_CALLBACK_PATH, by, common.EncodeID(int64(transferID)))
}

func SendReceiptUrl(transferID int, by string) string {
	return fmt.Sprintf("https://debtstracker.io/app/send-receipt?by=%v&transfer=%v", by, common.EncodeIntID(transferID))
}
