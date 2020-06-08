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
			"02 Jan 06 15:04 MST",
			"03 Jan 07 10:04 MST",
			"04 Jan 06 09:04 MST",
		},
		invalidTimeStrings: []string{},
	},
	{
		timePeriod: TimePeriod{
			dates: []InclusiveRange{{begin: 15, end: 15}},
		},
		validTimeStrings: []string{
			"15 Jan 06 15:04 MST",
			"15 Mar 07 10:04 MST",
			"15 Dec 06 09:04 MST",
		},
		invalidTimeStrings: []string{
			"14 Jan 06 15:04 MST",
			"16 Mar 07 10:04 MST",
			"14 Dec 06 23:59 MST",
		},
	},
}

func TestContainsTime(t *testing.T) {
	for _, tc := range timePeriodTestCases {
		for _, ts := range tc.validTimeStrings {
			_t, _ := time.Parse(time.RFC822, ts)
			if !tc.timePeriod.ContainsTime(_t) {
				t.Errorf("Expected period %+v to contain %+v", tc.timePeriod, _t)
			}
		}
		for _, ts := range tc.invalidTimeStrings {
			_t, _ := time.Parse(time.RFC822, ts)
			if tc.timePeriod.ContainsTime(_t) {
				t.Errorf("Period %+v not expected to contain %+v", tc.timePeriod, _t)
			}
		}
	}
}
