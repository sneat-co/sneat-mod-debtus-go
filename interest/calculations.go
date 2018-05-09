package interest

import (
	"github.com/strongo/decimal"
	"time"
	"fmt"
)

var calculators = map[Formula]Calculator{
	simplePercentCalculator.Formula(): simplePercentCalculator,
}

func Calculate(reportTime time.Time, deal Deal, payments []Payment) (interest, outstanding decimal.Decimal64p2, err error) {
	formula := deal.Formula()
	calculator := calculators[formula]
	if calculator == nil {
		err = fmt.Errorf("unknown formula (%v)", formula)
		return
	}
	return calculator.Calculate(reportTime, deal, payments)
}
