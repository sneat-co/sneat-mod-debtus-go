package interest

import (
	"github.com/strongo/decimal"
	"time"
	"fmt"
)

type SimplePercentCalculator struct {
}

var simplePercentCalculator = SimplePercentCalculator{}

var _ Calculator = (*SimplePercentCalculator)(nil)

func (SimplePercentCalculator) Formula() Formula {
	return "simple"
}

func (SimplePercentCalculator) Calculate(reportTime time.Time, deal Deal, payments []Payment) (interest, outstanding decimal.Decimal64p2, err error) {
	if reportTime.IsZero() {
		panic("report time should be non zero")
	}
	if reportTime.Before(deal.Time()) {
		panic("report time should be after deal time")
	}

	firstPeriod := true
	outstanding = deal.LentAmount()

	interestRatePerDay := deal.RatePercent().AsFloat64() / float64(deal.RatePeriod() * 100)

	calculatePeriod := func(starts, ends time.Time) () {
		if outstanding <= 0 {
			return // TODO: Raise panic if < 0?
		}
		var interestForPeriod decimal.Decimal64p2
		ageInDays := AgeInDays(starts, ends)
		if ageInDays < deal.MinimumPeriod() {
			ageInDays = deal.MinimumPeriod()
		}

		if firstPeriod {
			firstPeriod = false
		} else {
			ageInDays -= 1
		}
		interestForPeriod = decimal.NewDecimal64p2FromFloat64(outstanding.AsFloat64() * interestRatePerDay * float64(ageInDays))
		outstanding += interestForPeriod
		interest += interestForPeriod
		return
	}

	periodStarts := deal.Time()
	for i, payment := range payments {
		paymentTime := payment.Time()
		if paymentTime.IsZero() {
			err = fmt.Errorf("payment #%v has zero time", i)
			return
		}
		calculatePeriod(periodStarts, paymentTime)
		outstanding -= payment.Amount()
		periodStarts = paymentTime
	}
	calculatePeriod(periodStarts, reportTime)
	return
}
