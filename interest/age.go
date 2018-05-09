package interest

import "time"

func AgeInDays(periodStarts, periodEnds time.Time) int {
	hours := periodEnds.Sub(periodStarts).Hours()
	return int((hours + 24) / 24) // The day of debt issuing is counted as a whole day even if 1 second passed.
}


