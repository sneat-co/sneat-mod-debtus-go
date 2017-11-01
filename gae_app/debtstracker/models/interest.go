package models

import (
	"github.com/strongo/decimal"
	"fmt"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/pkg/errors"
)

type InterestRatePeriod int

const (
	InterestRatePeriodDaily   = 1
	InterestRatePeriodWeekly  = 7
	InterestRatePeriodMonthly = 30
	InterestRatePeriodYearly  = 360
)

type InterestPercentType string

const (
	InterestPercentSimple   InterestPercentType = "simple"
	InterestPercentCompound InterestPercentType = "compound"
)

type TransferInterest struct {
	InterestPeriod        InterestRatePeriod
	InterestPercent       decimal.Decimal64p2
	InterestType          InterestPercentType
	InterestAmountInCents decimal.Decimal64p2
}

var (
	ErrInterestTypeIsNotSet = errors.New("InterestType is not set")
)

func (entity TransferInterest) validateTransferInterest() (err error) {
	if entity.InterestPeriod == 0 && entity.InterestAmountInCents == 0 && entity.InterestPercent == 0 && entity.InterestType == "" {
		return
	}
	if entity.InterestPeriod < 0 {
		return fmt.Errorf("InterestPeriod < 0: %v", entity.InterestPeriod)
	}
	if entity.InterestPercent <= 0 {
		return fmt.Errorf("InterestPercent <= 0: %v", entity.InterestPercent)
	}
	if entity.InterestAmountInCents < 0 {
		return fmt.Errorf("InterestAmountInCents < 0: %v", entity.InterestAmountInCents)
	}
	if entity.InterestType == "" {
		return ErrInterestTypeIsNotSet
	}
	if entity.InterestType == InterestPercentSimple && entity.InterestType != InterestPercentCompound {
		return fmt.Errorf("unknown InterestType: %v", entity.InterestType)
	}
	if entity.InterestPeriod == 0 || entity.InterestAmountInCents == 0 || entity.InterestPercent == 0 {
		return fmt.Errorf(
			"one of values is 0: InterestPeriod=%v, InterestAmountInCents=%v, InterestPercent=%v",
			entity.InterestPeriod,
			entity.InterestAmountInCents,
			entity.InterestPercent,
		)
	}
	return
}

func (entity TransferInterest) cleanInterestProperties(properties []datastore.Property) ([]datastore.Property, error) {
	return gaedb.CleanProperties(properties, map[string]gaedb.IsOkToRemove{
		"InterestPeriod":        gaedb.IsZeroInt,
		"InterestPercent":       gaedb.IsZeroInt,
		"InterestType":          gaedb.IsEmptyString,
		"InterestAmountInCents": gaedb.IsZeroInt,
	})
}
