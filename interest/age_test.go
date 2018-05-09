package interest

import (
	"testing"
	"time"
)

func TestAgeInDays(t *testing.T) {
	for i, period := range []struct{
		name string
		starts time.Time
		ends time.Time
		expectedAgeInDays int
	}{
		{
			name: "same day + 1h",
			starts: time.Date(2010, 1, 1, 1, 1, 1, 0, time.UTC),
			ends: time.Date(2010, 1, 1, 2, 1, 1, 0, time.UTC),
			expectedAgeInDays: 1,
		},
		{
			name: "next day + 1h",
			starts: time.Date(2010, 1, 1, 1, 1, 1, 0, time.UTC),
			ends: time.Date(2010, 1, 2, 2, 1, 1, 0, time.UTC),
			expectedAgeInDays: 2,
		},
	} {
		if v := AgeInDays(period.starts, period.ends); v != period.expectedAgeInDays {
			t.Errorf("Case #%v - %v: %v != %v", i, period.name, period.expectedAgeInDays, v)
		}
	}
}
