package common

import (
	"fmt"
	"strings"
)

func GetReceiptUrl(receiptID int, host string) string {
	if receiptID == 0 {
		panic("receiptID == 0")
	}
	if host == "" {
		panic("host is empty string")
	} else if !strings.Contains(host, ".") {
		panic("host is not a domain name: " + host)
	}
	return fmt.Sprintf("https://%v/receipt?id=%v", host, EncodeIntID(receiptID))
}
