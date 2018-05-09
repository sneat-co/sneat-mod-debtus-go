package interest

import (
	"github.com/strongo/decimal"
	"time"
	"fmt"
)

type Credit interface {
	Formula() Formula
	RatePeriod() RatePeriodInDays
	RatePercent() decimal.Decimal64p2
	MinimumPeriod() int
	GracePeriod() int
}

type Deal interface {
	Credit
	Time() time.Time
	LentAmount() decimal.Decimal64p2
}

type deal struct {
	formula       Formula
	time          time.Time
	lentAmount    decimal.Decimal64p2
	ratePeriod    RatePeriodInDays
	ratePercent   decimal.Decimal64p2
	minimumPeriod int
	gracePeriod   int
}

func (d deal) Formula() Formula {
	return d.formula
}

func (d deal) LentAmount() decimal.Decimal64p2 {
	return d.lentAmount
}

func (d deal) RatePercent() decimal.Decimal64p2 {
	return d.ratePercent
}

func (d deal) RatePeriod() RatePeriodInDays {
	return d.ratePeriod
}

func (d deal) MinimumPeriod() int {
	return d.minimumPeriod
}

func (d deal) GracePeriod() int {
	return d.gracePeriod
}

func (d deal) Time() time.Time {
	return d.time
}

func NewDeal(formula Formula, time time.Time, lentAmount, ratePercent decimal.Decimal64p2, ratePeriod RatePeriodInDays, minimumPeriod, gracePeriod int) Deal {
	if lentAmount < 0 {
		panic(fmt.Sprintf("lentAmount <= 0: %v", lentAmount))
	}
	if ratePercent < 0 {
		panic(fmt.Sprintf("ratePercent <= 0: %v", ratePercent))
	}
	if minimumPeriod < 0 {
		panic(fmt.Sprintf("minimumPeriod <= 0: %v", minimumPeriod))
	}
	if gracePeriod < 0 {
		panic(fmt.Sprintf("gracePeriod <= 0: %v", gracePeriod))
	}
	if ratePeriod <= 0 {
		panic(fmt.Sprintf("ratePeriod <= 0: %v", ratePeriod))
	}
	return deal{
		formula:       formula,
		time:          time,
		lentAmount:    lentAmount,
		ratePeriod:    ratePeriod,
		ratePercent:   ratePercent,
		minimumPeriod: minimumPeriod,
		gracePeriod:   gracePeriod,
	}
}

type Payment interface {
	Time() time.Time
	Amount() decimal.Decimal64p2
}

type payment struct {
	time   time.Time
	amount decimal.Decimal64p2
}

func (p payment) Time() time.Time {
	return p.time
}

func (p payment) Amount() decimal.Decimal64p2 {
	return p.amount
}

func NewPayment(time time.Time, amount decimal.Decimal64p2) Payment {
	return payment{time: time, amount: amount}
}

type Calculator interface {
	Formula() Formula
	Calculate(reportTime time.Time, deal Deal, payments []Payment) (interest, outstanding decimal.Decimal64p2, err error)
}
