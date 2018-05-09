package interest

type Formula string

const (
	FormulaSimple   = "simple"
	FormulaCompound = "compound"
)

var KnownFormulas = []Formula{
	FormulaSimple,
	FormulaCompound,
}

type RatePeriodInDays int

const (
	RatePeriodDaily   = 1
	RatePeriodWeekly  = 7
	RatePeriodMonthly = 30
	RatePeriodYearly  = 360
)

func IsKnownFormula(formula Formula) bool {
	for _, f := range KnownFormulas {
		if f == formula {
			return true
		}
	}
	return false
}