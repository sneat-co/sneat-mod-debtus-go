package models

import (
	"github.com/strongo/decimal"
	"fmt"
	//"google.golang.org/appengine/datastore"
	//"github.com/strongo/app/gaedb"
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
	InterestPercentSimple   = "simple"
	InterestPercentCompound = "compound"
)

type TransferInterest struct {
	InterestType          InterestPercentType `datastore:",noindex,omitempty"`
	InterestPeriod        int                 `datastore:",noindex,omitempty"`
	InterestPercent       decimal.Decimal64p2 `datastore:",noindex,omitempty"`
	InterestGracePeriod   int                 `datastore:",noindex,omitempty" json:",omitempty"` // How many days are without any interest
	InterestMinimumPeriod int                 `datastore:",noindex,omitempty" json:",omitempty"` // Minimum days for interest (e.g. penalty for earlier return).
	//InterestAmountInCents decimal.Decimal64p2 `datastore:",noindex" json:",omitempty"`
}

func (t TransferInterest) HasInterest() bool {
	return t.InterestPercent != 0
}

func (t TransferInterest) GetInterestData() TransferInterest {
	return t
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
		if len(returns) != len(t.ReturnTransferIDs) && len(returns) > 0 {
			t.ReturnTransferIDs = nil
			//return fmt.Errorf("len(t.GetReturns()) != len(t.ReturnTransferIDs): %v != %v", len(t.GetReturns()), len(t.ReturnTransferIDs))
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

func (ti TransferInterest) validateTransferInterest() (err error) {
	if ti.InterestPeriod == 0 && ti.InterestPercent == 0 && ti.InterestType == "" {
		return
	}
	if ti.InterestPeriod < 0 {
		return fmt.Errorf("InterestPeriod < 0: %v", ti.InterestPeriod)
	}
	if ti.InterestPercent <= 0 {
		return fmt.Errorf("InterestPercent <= 0: %v", ti.InterestPercent)
	}
	//if entity.InterestAmountInCents < 0 {
	//	return fmt.Errorf("InterestAmountInCents < 0: %v", entity.InterestAmountInCents)
	//}
	if ti.InterestType == "" {
		return ErrInterestTypeIsNotSet
	}
	if ti.InterestType != InterestPercentSimple && ti.InterestType != InterestPercentCompound {
		return fmt.Errorf("unknown InterestType: %v", ti.InterestType)
	}
	if ti.InterestPeriod == 0 || ti.InterestPercent == 0 {
		return fmt.Errorf(
			"one of values is 0: InterestPeriod=%v, InterestPercent=%v",
			ti.InterestPeriod,
			ti.InterestPercent,
		)
	}
	return
}

//func init() {
//	addInterestPropertiesToClean := func(props2clean map[string]gaedb.IsOkToRemove) {
//		props2clean["InterestType"] = gaedb.IsEmptyString
//		props2clean["InterestPeriod"] = gaedb.IsZeroInt
//		props2clean["InterestPercent"] = gaedb.IsZeroInt
//		props2clean["InterestGracePeriod"] = gaedb.IsZeroInt
//		props2clean["InterestMinimumPeriod"] = gaedb.IsZeroInt
//	}
//	addInterestPropertiesToClean(transferPropertiesToClean)
//}

func (t *TransferEntity) GetOutstandingValue() (outstandingValue decimal.Decimal64p2) {
	if t.IsReturn && t.AmountInCentsReturned == 0 {
		return 0
	}
	interestValue := t.GetInterestValue()
	outstandingValue = t.AmountInCents + interestValue - t.AmountInCentsReturned
	if outstandingValue < 0 {
		panic(fmt.Sprintf("outstandingValue < 0: %v, IsReturn: %v, Amount: %v, Returned: %v, Interest: %v", outstandingValue, t.IsReturn, t.AmountInCents, t.AmountInCentsReturned, t.GetInterestValue()))
	}
	return
}


func (t *TransferEntity) GetOutstandingAmount() Amount {
	return Amount{Currency: t.Currency, Value: t.GetOutstandingValue()}
}

type TransferInterestCalculable interface {
	GetLendingValue() decimal.Decimal64p2
	GetStartDate() time.Time
	GetReturns() []TransferReturnJson
	GetInterestData() TransferInterest
}

func (t *TransferEntity) GetInterestValue() (interestValue decimal.Decimal64p2) {
	return CalculateInterestValue(t)
}

func CalculateInterestValue(t TransferInterestCalculable) (interestValue decimal.Decimal64p2) {
	firstPeriod := true
	interestData := t.GetInterestData()
	outstanding := t.GetLendingValue()
	getSimpleInterestForPeriod := func(starts, ends time.Time) (interestAmount decimal.Decimal64p2) {
		if outstanding <= 0 {
			return 0
		}
		interestRate := interestData.InterestPercent.AsFloat64() / 100
		interestRatePerDay := interestRate / float64(interestData.InterestPeriod)
		ageInDays := ageInDays(starts, ends)
		if ageInDays < interestData.InterestMinimumPeriod {
			ageInDays = interestData.InterestMinimumPeriod
		}

		if firstPeriod {
			firstPeriod = false
		} else {
			ageInDays -= 1
		}
		interestAmount = decimal.NewDecimal64p2FromFloat64(outstanding.AsFloat64() * interestRatePerDay * float64(ageInDays))
		return
	}
	switch interestData.InterestType {
	case InterestPercentSimple:
		periodStarts := t.GetStartDate()
		for _, transferReturn := range t.GetReturns() {
			interestValue += getSimpleInterestForPeriod(periodStarts, transferReturn.Time)
			outstanding -= transferReturn.Amount
			periodStarts = transferReturn.Time
		}
		periodEnds := time.Now()
		interestValue += getSimpleInterestForPeriod(periodStarts, periodEnds)
	case InterestPercentCompound:
		panic("not implemented")
	case "":
		// Ignore
	default:
		panic(fmt.Sprintf("unknown interest type: %v", interestData.InterestType))
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
