package common

import "fmt"

func GetReceiptUrl(receiptID int64, host string) string {
	if receiptID == 0 {
		panic("receiptID == 0")
	}
	if host == "" {
		panic("host is empty string")
	}
	return fmt.Sprintf("https://%v/receipt?id=%v", host, EncodeID(receiptID))
}
