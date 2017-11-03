package models

import (
	"github.com/strongo/decimal"
	"fmt"
	"google.golang.org/appengine/datastore"
	"github.com/strongo/app/gaedb"
	"github.com/pkg/errors"
	"time"
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
	InterestType          InterestPercentType `datastore:",noindex"`
	InterestPeriod        InterestRatePeriod  `datastore:",noindex"`
	InterestPercent       decimal.Decimal64p2 `datastore:",noindex"`
	InterestGracePeriod   int                 `datastore:",noindex"` // How many days are without any interest
	InterestMinimumPeriod int                 `datastore:",noindex"` // Minimum days for interest (e.g. penalty for earlier return).
	InterestAmountInCents decimal.Decimal64p2 `datastore:",noindex"`
}

var (
	ErrInterestTypeIsNotSet = errors.New("InterestType is not set")
)

func (t *TransferEntity) validateTransferInterestAndReturns() (err error) {
	if err = t.TransferInterest.validateTransferInterest(); err != nil {
		return
	}
	if t.InterestType != "" { // TODO: Migrate old records and then do the check for all transfers
		returns := t.GetReturns()
		if len(returns) != len(t.ReturnTransferIDs) {
			return fmt.Errorf("len(t.GetReturns()) != len(t.ReturnTransferIDs): %v != %v", len(t.GetReturns()), len(t.ReturnTransferIDs))
		}
		var amountReturned decimal.Decimal64p2
		for _, r := range returns {
			amountReturned += r.Amount
		}
		if amountReturned != t.AmountInCentsReturned {
			return fmt.Errorf("sum(returns.Amount) != *TransferEntity.AmountInCentsReturned: %v != %v", amountReturned, t.AmountInCentsReturned)
		}
	}
	return
}

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
		"InterestType":          gaedb.IsEmptyString,
		"InterestPeriod":        gaedb.IsZeroInt,
		"InterestPercent":       gaedb.IsZeroInt,
		"InterestGracePeriod":   gaedb.IsZeroInt,
		"InterestMinimumPeriod": gaedb.IsZeroInt,
		"InterestAmountInCents": gaedb.IsZeroInt,
	})
}

func (t *TransferEntity) GetOutstandingAmount() Amount {
	if t.InterestType == "" {
		return Amount{Currency: t.Currency, Value: t.AmountInCentsOutstanding}
	}
	outstandingValue := t.AmountInCents - t.AmountInCentsReturned + t.GetInterestAmount()
	return Amount{Currency: t.Currency, Value: outstandingValue}
}

func (t *TransferEntity) GetInterestAmount() (interestAmount decimal.Decimal64p2) {
	firstPeriod := true
	outstanding := t.AmountInCents
	getSimpleInterestForPeriod := func(starts, ends time.Time) (interestAmount decimal.Decimal64p2) {
		var interestRatePerDay = t.InterestPercent.AsFloat64() / float64(t.InterestPeriod) / 100
		ageInDays := ageInDays(starts, ends)
		if ageInDays < t.InterestMinimumPeriod {
			ageInDays = t.InterestMinimumPeriod
		}

		if firstPeriod {
			firstPeriod = false
		} else {
			ageInDays -= 1
		}
		interestAmount = decimal.NewDecimal64p2FromFloat64(float64(outstanding) * interestRatePerDay * float64(ageInDays))
		return
	}
	switch t.InterestType {
	case InterestPercentSimple:
		periodStarts := t.DtCreated
		for _, transferReturn := range t.GetReturns() {
			interestAmount += getSimpleInterestForPeriod(periodStarts, transferReturn.Time)
			outstanding -= transferReturn.Amount
			periodStarts = transferReturn.Time
		}
		periodEnds := time.Now()
		interestAmount += getSimpleInterestForPeriod(periodStarts, periodEnds)
	case InterestPercentCompound:
		panic("not implemented")
	case "":
		// Ignore
	default:
		panic(fmt.Sprintf("unknown interest type: %v", t.InterestType))
	}
	return
}

func ageInDays(periodStarts, periodEnds time.Time) int {
	hours := periodEnds.Sub(periodStarts).Hours()
	return int((hours + 24) / 24) // The day of debt issuing is counted as a whole day even if 1 second passed.
}

func (t *TransferEntity) AgeInDays() int {
	return ageInDays(time.Now(), t.DtCreated)
}

/*
Example:

7% per week min 3 days
1.5% в неделю мин 3 дня

 */
