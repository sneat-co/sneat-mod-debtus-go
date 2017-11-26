package models

import (
	"fmt"

	"github.com/strongo/decimal"
	//"google.golang.org/appengine/datastore"
	//"github.com/strongo/db/gaedb"
	"time"

	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
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
	if t.AmountInCentsInterest < 0 {
		panic(fmt.Sprintf("t.AmountInCentsInterest < 0: %v", t.AmountInCentsInterest))
	}
	if !t.IsReturn && t.AmountInCentsInterest != 0 {
		panic(fmt.Sprintf("!t.IsReturn && t.AmountInCentsInterest != 0: %v", t.AmountInCentsInterest))
	}
	if t.AmountInCentsInterest > t.AmountInCents {
		panic(fmt.Sprintf("t.AmountInCentsInterest > t.AmountInCents: %v > %v", t.AmountInCentsInterest, t.AmountInCents))
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

func (t *TransferEntity) GetOutstandingValue(periodEnds time.Time) (outstandingValue decimal.Decimal64p2) {
	if t.IsReturn && t.AmountInCentsReturned == 0 {
		return 0
	}
	interestValue := t.GetInterestValue(periodEnds)
	outstandingValue = t.AmountInCents + interestValue - t.AmountInCentsReturned
	if outstandingValue < 0 && interestValue != 0 {
		panic(fmt.Sprintf("outstandingValue < 0: %v, IsReturn: %v, Amount: %v, Returned: %v, Interest: %v\n%v", outstandingValue, t.IsReturn, t.AmountInCents, t.AmountInCentsReturned, interestValue, litter.Sdump(t)))
	}
	return
}

func (t *TransferEntity) GetOutstandingAmount(periodEnds time.Time) Amount {
	return Amount{Currency: t.Currency, Value: t.GetOutstandingValue(periodEnds)}
}

type TransferInterestCalculable interface {
	GetLendingValue() decimal.Decimal64p2
	GetStartDate() time.Time
	GetReturns() []TransferReturnJson
	GetInterestData() TransferInterest
}

func (t *TransferEntity) GetInterestValue(periodEnds time.Time) (interestValue decimal.Decimal64p2) {
	return CalculateInterestValue(t, periodEnds)
}

func CalculateInterestValue(t TransferInterestCalculable, periodEnds time.Time) (interestValue decimal.Decimal64p2) {
	firstPeriod := true
	interestData := t.GetInterestData()
	outstanding := t.GetLendingValue()
	calculateSimpleInterest := func() (interestAmount decimal.Decimal64p2) {
		interestRate := interestData.InterestPercent.AsFloat64() / 100
		interestRatePerDay := interestRate / float64(interestData.InterestPeriod)

		getSimpleInterestForPeriod := func(starts, ends time.Time) (simpleInterest decimal.Decimal64p2) {
			if outstanding <= 0 {
				return 0
			}
			ageInDays := ageInDays(starts, ends)
			if ageInDays < interestData.InterestMinimumPeriod {
				ageInDays = interestData.InterestMinimumPeriod
			}

			if firstPeriod {
				firstPeriod = false
			} else {
				ageInDays -= 1
			}
			simpleInterest = decimal.NewDecimal64p2FromFloat64(outstanding.AsFloat64() * interestRatePerDay * float64(ageInDays))
			return
		}

		periodStarts := t.GetStartDate()
		for _, transferReturn := range t.GetReturns() {
			interestForPeriod := getSimpleInterestForPeriod(periodStarts, transferReturn.Time)
			if transferReturn.Amount < interestForPeriod {
				unpaidInterest := interestForPeriod - transferReturn.Amount
				interestValue += unpaidInterest
				outstanding += unpaidInterest
			}
			outstanding -= transferReturn.Amount
			periodStarts = transferReturn.Time
		}
		interestValue += getSimpleInterestForPeriod(periodStarts, periodEnds)
		return
	}

	switch interestData.InterestType {
	case InterestPercentSimple:
		calculateSimpleInterest()
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

func updateBalanceWithInterest(b Balance, outstandingWithInterest []TransferWithInterestJson, periodEnds time.Time) {
	for _, outstandingTransferWithInterest := range outstandingWithInterest {
		balanceValue, ok := b[outstandingTransferWithInterest.Currency]
		if ok {
			interestValue := CalculateInterestValue(outstandingTransferWithInterest, periodEnds)
			if balanceValue < 0 {
				interestValue = -interestValue
			}
			b[outstandingTransferWithInterest.Currency] = balanceValue + interestValue
		} else {
			panic(fmt.Errorf("outstanding transfer %v with currency %v is not presented in balance", outstandingTransferWithInterest.TransferID, outstandingTransferWithInterest.Currency))
		}
	}
}

/*
Example:

7% per week min 3 days
1.5% в неделю мин 3 дня

*/
