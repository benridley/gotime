package gotime

import (
	"reflect"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

var timeIntervalTestCases = []struct {
	validTimeStrings   []string
	invalidTimeStrings []string
	timeInterval       TimeInterval
}{
	{
		timeInterval: TimeInterval{},
		validTimeStrings: []string{
			"02 Jan 06 15:04 MST",
			"03 Jan 07 10:04 MST",
			"04 Jan 06 09:04 MST",
		},
		invalidTimeStrings: []string{},
	},
	{
		// 9am to 5pm, monday to friday
		timeInterval: TimeInterval{
			Times:    []TimeRange{{StartMinute: 540, EndMinute: 1020}},
			Weekdays: []WeekdayRange{{InclusiveRange{Begin: 1, End: 5}}},
		},
		validTimeStrings: []string{
			"04 May 20 15:04 MST",
			"05 May 20 10:04 MST",
			"09 Jun 20 09:04 MST",
		},
		invalidTimeStrings: []string{
			"03 May 20 15:04 MST",
			"04 May 20 08:59 MST",
			"05 May 20 05:00 MST",
		},
	},
	{
		// Easter 2020
		timeInterval: TimeInterval{
			DaysOfMonth: []DayOfMonthRange{{InclusiveRange{Begin: 4, End: 6}}},
			Months:      []MonthRange{{InclusiveRange{Begin: 4, End: 4}}},
			Years:       []YearRange{{InclusiveRange{Begin: 2020, End: 2020}}},
		},
		validTimeStrings: []string{
			"04 Apr 20 15:04 MST",
			"05 Apr 20 00:00 MST",
			"06 Apr 20 23:05 MST",
		},
		invalidTimeStrings: []string{
			"03 May 18 15:04 MST",
			"03 Apr 20 23:59 MST",
			"04 Jun 20 23:59 MST",
			"06 Apr 19 23:59 MST",
			"07 Apr 20 00:00 MST",
		},
	},
	{
		// Check negative days of month, last 3 days of each month
		timeInterval: TimeInterval{
			DaysOfMonth: []DayOfMonthRange{{InclusiveRange{Begin: -3, End: -1}}},
		},
		validTimeStrings: []string{
			"31 Jan 20 15:04 MST",
			"30 Jan 20 15:04 MST",
			"29 Jan 20 15:04 MST",
			"30 Jun 20 00:00 MST",
			"29 Feb 20 23:05 MST",
		},
		invalidTimeStrings: []string{
			"03 May 18 15:04 MST",
			"27 Jan 20 15:04 MST",
			"03 Apr 20 23:59 MST",
			"04 Jun 20 23:59 MST",
			"06 Apr 19 23:59 MST",
			"07 Apr 20 00:00 MST",
			"01 Mar 20 00:00 MST",
		},
	},
	{
		// Check out of bound days are clamped to month boundaries
		timeInterval: TimeInterval{
			Months:      []MonthRange{{InclusiveRange{Begin: 6, End: 6}}},
			DaysOfMonth: []DayOfMonthRange{{InclusiveRange{Begin: -31, End: 31}}},
		},
		validTimeStrings: []string{
			"30 Jun 20 00:00 MST",
			"01 Jun 20 00:00 MST",
		},
		invalidTimeStrings: []string{
			"31 May 20 00:00 MST",
			"1 Jul 20 00:00 MST",
		},
	},
}

var timeStringTestCases = []struct {
	timeString  string
	TimeRange   TimeRange
	expectError bool
}{
	{
		timeString:  "{'start_time': '00:00', 'end_time': '24:00'}",
		TimeRange:   TimeRange{StartMinute: 0, EndMinute: 1440},
		expectError: false,
	},
	{
		timeString:  "{'start_time': '01:35', 'end_time': '17:39'}",
		TimeRange:   TimeRange{StartMinute: 95, EndMinute: 1059},
		expectError: false,
	},
	{
		timeString:  "{'start_time': '09:35', 'end_time': '09:39'}",
		TimeRange:   TimeRange{StartMinute: 575, EndMinute: 579},
		expectError: false,
	},
	{
		// Error: Begin and End times are the same
		timeString:  "{'start_time': '17:31', 'end_time': '17:31'}",
		TimeRange:   TimeRange{},
		expectError: true,
	},
	{
		// Error: End time out of range
		timeString:  "{'start_time': '12:30', 'end_time': '24:01'}",
		TimeRange:   TimeRange{},
		expectError: true,
	},
	{
		// Error: Start time greater than End time
		timeString:  "{'start_time': '09:30', 'end_time': '07:41'}",
		TimeRange:   TimeRange{},
		expectError: true,
	},
	{
		// Error: Start time out of range and greater than End time
		timeString:  "{'start_time': '24:00', 'end_time': '17:41'}",
		TimeRange:   TimeRange{},
		expectError: true,
	},
	{
		// Error: No range specified
		timeString:  "{'start_time': '14:03'}",
		TimeRange:   TimeRange{},
		expectError: true,
	},
}

var dayOfWeekStringTestCases = []struct {
	dowString   string
	ranges      []WeekdayRange
	expectError bool
}{
	{
		dowString:   "['monday:friday', 'saturday']",
		ranges:      []WeekdayRange{{InclusiveRange{Begin: 1, End: 5}}, {InclusiveRange{Begin: 6, End: 6}}},
		expectError: false,
	},
}

var yamlUnmarshalTestCases = []struct {
	in          string
	intervals   []TimeInterval
	contains    []string
	excludes    []string
	expectError bool
}{
	{
		// Simple business hours test
		in: `
---
- weekdays: ['monday:friday']
  times:
    - start_time: '09:00'
      end_time: '17:00'
`,
		intervals: []TimeInterval{
			{
				Weekdays: []WeekdayRange{{InclusiveRange{Begin: 1, End: 5}}},
				Times:    []TimeRange{{StartMinute: 540, EndMinute: 1020}},
			},
		},
		contains: []string{
			"08 Jul 20 09:00 MST",
			"08 Jul 20 16:59 MST",
		},
		excludes: []string{
			"08 Jul 20 05:00 MST",
			"08 Jul 20 08:59 MST",
		},
		expectError: false,
	},
	{
		// More advanced test with negative indices and ranges
		in: `
---
  # Last week, excluding Saturday, of the first quarter of the year during business hours from 2020 to 2025 and 2030-2035
- weekdays: ['monday:friday', 'sunday']
  months: ['january:march']
  days_of_month: ['-7:-1']
  years: ['2020:2025', '2030:2035']
  times:
    - start_time: '09:00'
      end_time: '17:00'
`,
		intervals: []TimeInterval{
			{
				Weekdays:    []WeekdayRange{{InclusiveRange{Begin: 1, End: 5}}, {InclusiveRange{Begin: 0, End: 0}}},
				Times:       []TimeRange{{StartMinute: 540, EndMinute: 1020}},
				Months:      []MonthRange{{InclusiveRange{1, 3}}},
				DaysOfMonth: []DayOfMonthRange{{InclusiveRange{-7, -1}}},
				Years:       []YearRange{{InclusiveRange{2020, 2025}}, {InclusiveRange{2030, 2035}}},
			},
		},
		contains: []string{
			"27 Jan 21 09:00 MST",
			"28 Jan 21 16:59 MST",
			"29 Jan 21 13:00 MST",
			"31 Mar 25 13:00 MST",
			"31 Mar 25 13:00 MST",
			"31 Jan 35 13:00 MST",
		},
		excludes: []string{
			"30 Jan 21 13:00 MST", // Saturday
			"01 Apr 21 13:00 MST", // 4th month
			"30 Jan 26 13:00 MST", // 2026
			"31 Jan 35 17:01 MST", // After 5pm
		},
		expectError: false,
	},
	{
		// Start day before End day
		in: `
---
- weekdays: ['friday:monday']`,
		expectError: true,
	},
	{
		// Invalid weekdays
		in: `
---
- weekdays: ['blurgsday:flurgsday']
`,
		expectError: true,
	},
	{
		// 0 day of month
		in: `
---
- days_of_month: ['0']
`,
		expectError: true,
	},
	{
		// Too early day of month
		in: `
---
- days_of_month: ['-50:-20']
`,
		expectError: true,
	},
	{
		// Negative indices should work
		in: `
---
- days_of_month: ['1:-1']
`,
		intervals: []TimeInterval{
			{
				DaysOfMonth: []DayOfMonthRange{{InclusiveRange{1, -1}}},
			},
		},
		expectError: false,
	},
	{
		// Negative start date before positive End date
		in: `
---
- days_of_month: ['-15:5']
`,
		expectError: true,
	},
	{
		// Negative End date before positive postive start date
		in: `
---
- days_of_month: ['10:-25']
`,
		expectError: true,
	},
}

func TestYamlUnmarshal(t *testing.T) {
	for _, tc := range yamlUnmarshalTestCases {
		var ti []TimeInterval
		err := yaml.Unmarshal([]byte(tc.in), &ti)
		if err != nil && !tc.expectError {
			t.Errorf("Received unexpected error: %v when parsing %v", err, tc.in)
		} else if err == nil && tc.expectError {
			t.Errorf("Expected error when unmarshalling %s but didn't receive one", tc.in)
		} else if err != nil && tc.expectError {
			continue
		}
		if !reflect.DeepEqual(ti, tc.intervals) {
			t.Errorf("Error unmarshalling %s: Want %+v, got %+v", tc.in, tc.intervals, ti)
		}
		for _, ts := range tc.contains {
			_t, _ := time.Parse(time.RFC822, ts)
			isContained := false
			for _, interval := range ti {
				if interval.ContainsTime(_t) {
					isContained = true
				}
			}
			if !isContained {
				t.Errorf("Expected intervals to contain time %s", _t)
			}
		}
		for _, ts := range tc.excludes {
			_t, _ := time.Parse(time.RFC822, ts)
			isContained := false
			for _, interval := range ti {
				if interval.ContainsTime(_t) {
					isContained = true
				}
			}
			if isContained {
				t.Errorf("Expected intervals to exclude time %s", _t)
			}
		}
	}
}

func TestContainsTime(t *testing.T) {
	for _, tc := range timeIntervalTestCases {
		for _, ts := range tc.validTimeStrings {
			_t, _ := time.Parse(time.RFC822, ts)
			if !tc.timeInterval.ContainsTime(_t) {
				t.Errorf("Expected period %+v to contain %+v", tc.timeInterval, _t)
			}
		}
		for _, ts := range tc.invalidTimeStrings {
			_t, _ := time.Parse(time.RFC822, ts)
			if tc.timeInterval.ContainsTime(_t) {
				t.Errorf("Period %+v not expected to contain %+v", tc.timeInterval, _t)
			}
		}
	}
}

func TestParseTimeString(t *testing.T) {
	for _, tc := range timeStringTestCases {
		var tr TimeRange
		err := yaml.Unmarshal([]byte(tc.timeString), &tr)
		if err != nil && !tc.expectError {
			t.Errorf("Received unexpected error: %v when parsing %v", err, tc.timeString)
		} else if err == nil && tc.expectError {
			t.Errorf("Expected error for invalid string %s but didn't receive one", tc.timeString)
		} else if !reflect.DeepEqual(tr, tc.TimeRange) {
			t.Errorf("Error parsing time string %s: Want %+v, got %+v", tc.timeString, tc.TimeRange, tr)
		}
	}
}

func TestParseWeek(t *testing.T) {
	for _, tc := range dayOfWeekStringTestCases {
		var wr []WeekdayRange
		err := yaml.Unmarshal([]byte(tc.dowString), &wr)
		if err != nil && !tc.expectError {
			t.Errorf("Received unexpected error: %v when parsing %v", err, tc.dowString)
		} else if err == nil && tc.expectError {
			t.Errorf("Expected error for invalid string %s but didn't receive one", tc.dowString)
		} else if !reflect.DeepEqual(wr, tc.ranges) {
			t.Errorf("Error parsing time string %s: Want %+v, got %+v", tc.dowString, tc.ranges, wr)
		}
	}
}

func TestYamlMarshal(t *testing.T) {
	for _, tc := range yamlUnmarshalTestCases {
		if tc.expectError {
			continue
		}
		var ti []TimeInterval
		err := yaml.Unmarshal([]byte(tc.in), &ti)
		if err != nil {
			t.Error(err)
		}
		out, err := yaml.Marshal(&ti)
		if err != nil {
			t.Error(err)
		}
		var ti2 []TimeInterval
		yaml.Unmarshal(out, &ti2)
		if !reflect.DeepEqual(ti, ti2) {
			t.Errorf("Re-marshalling %s produced a different TimeInterval", tc.in)
		}
	}
}

func emptyInterval() TimeInterval {
	return TimeInterval{
		Times:       []TimeRange{},
		Weekdays:    []WeekdayRange{},
		DaysOfMonth: []DayOfMonthRange{},
		Months:      []MonthRange{},
		Years:       []YearRange{},
	}
}
