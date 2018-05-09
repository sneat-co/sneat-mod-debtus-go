package interest

import (
	"testing"
	"github.com/strongo/decimal"
	"time"
)

func TestSimplePercentCalculator_Formula(t *testing.T) {
	calculator := SimplePercentCalculator{}
	if formula := calculator.Formula(); formula != "simple" {
		t.Fatalf("name should be 'simple', got: %v", formula)
	}
}

func TestSimplePercentCalculator_Calculate(t *testing.T) {
	calculator := SimplePercentCalculator{}
	var deal Deal

	for i, testCase := range []struct {
		name                string
		dealTime            time.Time
		reportTime          time.Time
		RatePeriodInDays
		lentAmount          decimal.Decimal64p2
		ratePercent         decimal.Decimal64p2
		expectedInterest    decimal.Decimal64p2
		expectedOutstanding decimal.Decimal64p2
		payments            []Payment
	}{
		{
			name:                "same day report, no minimum or grace period and no payments",
			dealTime:            time.Date(2010, 1, 1, 1, 1, 1, 0, time.UTC),
			reportTime:          time.Date(2010, 1, 1, 2, 1, 1, 0, time.UTC),
			lentAmount:          decimal.FromInt(200),
			RatePeriodInDays:    RatePeriodDaily,
			ratePercent:         decimal.FromInt(3),
			expectedInterest:    decimal.NewDecimal64p2FromFloat64(6),
			expectedOutstanding: decimal.NewDecimal64p2FromFloat64(206),
		},
		{
			name:             "same day report, no minimum or grace period and full payments",
			dealTime:         time.Date(2010, 1, 1, 1, 1, 1, 0, time.UTC),
			reportTime:       time.Date(2010, 1, 1, 2, 1, 1, 0, time.UTC),
			lentAmount:       decimal.FromInt(200),
			RatePeriodInDays: RatePeriodDaily,
			ratePercent:      decimal.FromInt(3),
			payments: []Payment{
				NewPayment(time.Date(2010, 1, 1, 1, 2, 1, 0, time.UTC), decimal.NewDecimal64p2FromFloat64(206.00)),
			},
			expectedInterest:    decimal.NewDecimal64p2FromFloat64(6.00),
			expectedOutstanding: 0,
		},
	} {
		deal = NewDeal(FormulaSimple, testCase.dealTime, testCase.lentAmount, testCase.ratePercent, testCase.RatePeriodInDays, 0, 0)
		interest, outstanding, err := calculator.Calculate(testCase.reportTime, deal, testCase.payments)
		if err != nil {
			t.Errorf("Case #%v - %v: unexpected error: %v)", i, testCase.name, err)
			continue
		}
		if interest != testCase.expectedInterest {
			t.Errorf("Case #%v - %v: unexpected interest value (expected: %v, got: %v)",
				i, testCase.name, testCase.expectedInterest, interest)
		}
		if outstanding != testCase.expectedOutstanding {
			t.Errorf("Case #%v - %v: unexpected outstanding value (expected: %v, got: %v)",
				i, testCase.name, testCase.expectedOutstanding, outstanding)
		}
	}
}
