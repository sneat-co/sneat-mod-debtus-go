package models

import (
	"github.com/strongo/decimal"
	"fmt"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
)

type InterestRatePeriod int

const (
	InterestRatePeriodDaily   = 1
	InterestRatePeriodWeekly  = 7
	InterestRatePeriodMonthly = 30
	InterestRatePeriodYearly  = 365
)

type TransferInterest struct {
	InterestRatePeriod     InterestRatePeriod
	InterestRateInPercents decimal.Decimal64p2
	InterestAmountInCents  decimal.Decimal64p2
}

func (entity TransferInterest) ValidateTransferInterest() (err error) {
	if entity.InterestRatePeriod == 0 && entity.InterestAmountInCents == 0 && entity.InterestRateInPercents == 0 {
		return
	}
	if entity.InterestRatePeriod < 0 {
		return fmt.Errorf("entity.InterestRatePeriod < 0: %v", entity.InterestRatePeriod)
	}
	if entity.InterestRateInPercents < 0 {
		return fmt.Errorf("entity.InterestRateInPercents < 0: %v", entity.InterestRateInPercents)
	}
	if entity.InterestAmountInCents < 0 {
		return fmt.Errorf("entity.InterestAmountInCents < 0: %v", entity.InterestAmountInCents)
	}
	if entity.InterestRatePeriod == 0 || entity.InterestAmountInCents == 0 || entity.InterestRateInPercents == 0 {
		return fmt.Errorf(
			"one of values is 0: InterestRatePeriod=%v, InterestAmountInCents=%v, InterestRateInPercents=%v",
			entity.InterestRatePeriod,
			entity.InterestAmountInCents,
			entity.InterestRateInPercents,
		)
	}
	return
}

func (entity TransferInterest) cleanInterestProperties(properties []datastore.Property) ([]datastore.Property, error) {
	return gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"InterestRatePeriod": gaedb.IsZeroInt,
		"InterestRateInPercents": gaedb.IsZeroInt,
		"InterestAmountInCents": gaedb.IsZeroInt,
	})
}