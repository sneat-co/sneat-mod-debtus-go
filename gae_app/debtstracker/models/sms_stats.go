package models

import "github.com/strongo/decimal"

type SmsStats struct {
	SmsCount   int64   `datastore:",noindex"`
	SmsCost    float32 `datastore:",noindex"`
	SmsCostUSD decimal.Decimal64p2
}
