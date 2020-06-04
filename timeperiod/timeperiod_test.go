package timeperiod

import (
	"testing"
	"time"
)

var timePeriodTestCases = []struct {
	validTimeStrings   []string
	invalidTimeStrings []string
	timePeriod         TimePeriod
}{
	{
		timePeriod: TimePeriod{},
		validTimeStrings: []string{
			"Dec 11 03:04:05 -0700 MST 2006",
			"Jan 2 15:04:05 -0700 MST 2012",
			"Apr 1 23:59:59 -0700 MST 1999",
		},
		invalidTimeStrings: []string{},
	},
	{
		timePeriod: TimePeriod{
			dates: []InclusiveRange{{begin: 15, end: 15}},
		},
		validTimeStrings: []string{
			"Dec 11 03:04:05 -0700 MST 2006",
			"Jan 2 15:04:05 -0700 MST 2012",
			"Apr 1 23:59:59 -0700 MST 1999",
		},
		invalidTimeStrings: []string{},
	},
}

func TestContainsTime(t *testing.T) {
	for _, tc := range timePeriodTestCases {
		for _, ts := range tc.validTimeStrings {
			_t, _ := time.Parse("Jan 02 15:04:05 -0700 2006", ts)
			if !tc.timePeriod.ContainsTime(_t) {
				t.Errorf("Expected period %+v to contain %+v", tc.timePeriod, _t)
			}
		}
	}
}
