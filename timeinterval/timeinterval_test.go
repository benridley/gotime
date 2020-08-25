package timeinterval

import (
	"io/ioutil"
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
			Times:    []timeRange{{startMinute: 540, endMinute: 1020}},
			Weekdays: []weekdayRange{{inclusiveRange{begin: 1, end: 5}}},
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
			DaysOfMonth: []dayOfMonthRange{{inclusiveRange{begin: 4, end: 6}}},
			Months:      []monthRange{{inclusiveRange{begin: 4, end: 4}}},
			Years:       []yearRange{{inclusiveRange{begin: 2020, end: 2020}}},
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
			DaysOfMonth: []dayOfMonthRange{{inclusiveRange{begin: -3, end: -1}}},
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
			Months:      []monthRange{{inclusiveRange{begin: 6, end: 6}}},
			DaysOfMonth: []dayOfMonthRange{{inclusiveRange{begin: -31, end: 31}}},
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
	timeRange   timeRange
	expectError bool
}{
	{
		timeString:  "{'start_time': '00:00', 'end_time': '24:00'}",
		timeRange:   timeRange{startMinute: 0, endMinute: 1440},
		expectError: false,
	},
	{
		timeString:  "{'start_time': '01:35', 'end_time': '17:39'}",
		timeRange:   timeRange{startMinute: 95, endMinute: 1059},
		expectError: false,
	},
	{
		timeString:  "{'start_time': '09:35', 'end_time': '09:39'}",
		timeRange:   timeRange{startMinute: 575, endMinute: 579},
		expectError: false,
	},
	{
		// Error: begin and end times are the same
		timeString:  "{'start_time': '17:31', 'end_time': '17:31'}",
		timeRange:   timeRange{},
		expectError: true,
	},
	{
		// Error: end time out of range
		timeString:  "{'start_time': '12:30', 'end_time': '24:01'}",
		timeRange:   timeRange{},
		expectError: true,
	},
	{
		// Error: Start time greater than end time
		timeString:  "{'start_time': '09:30', 'end_time': '07:41'}",
		timeRange:   timeRange{},
		expectError: true,
	},
	{
		// Error: Start time out of range and greater than end time
		timeString:  "{'start_time': '24:00', 'end_time': '17:41'}",
		timeRange:   timeRange{},
		expectError: true,
	},
	{
		// Error: No range specified
		timeString:  "{'start_time': '14:03'}",
		timeRange:   timeRange{},
		expectError: true,
	},
}

var dayOfWeekStringTestCases = []struct {
	dowString   string
	ranges      []weekdayRange
	expectError bool
}{
	{
		dowString:   "['monday:friday', 'saturday']",
		ranges:      []weekdayRange{{inclusiveRange{begin: 1, end: 5}}, {inclusiveRange{begin: 6, end: 6}}},
		expectError: false,
	},
}

var yamlUnmarshalTestCases = []struct {
	yamlPath    string
	intervals   []TimeInterval
	contains    []string
	excludes    []string
	expectError bool
}{
	{
		// Simple business hours test
		yamlPath: "./tests/basic.yml",
		intervals: []TimeInterval{
			{
				Weekdays: []weekdayRange{{inclusiveRange{begin: 1, end: 5}}},
				Times:    []timeRange{{startMinute: 540, endMinute: 1020}},
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
		yamlPath: "./tests/advanced.yml",
		intervals: []TimeInterval{
			{
				Weekdays:    []weekdayRange{{inclusiveRange{begin: 1, end: 5}}, {inclusiveRange{begin: 0, end: 0}}},
				Times:       []timeRange{{startMinute: 540, endMinute: 1020}},
				Months:      []monthRange{{inclusiveRange{1, 3}}},
				DaysOfMonth: []dayOfMonthRange{{inclusiveRange{-7, -1}}},
				Years:       []yearRange{{inclusiveRange{2020, 2025}}, {inclusiveRange{2030, 2035}}},
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
}

func TestYamlUnmarshal(t *testing.T) {
	for _, tc := range yamlUnmarshalTestCases {
		f, err := ioutil.ReadFile(tc.yamlPath)
		if err != nil {
			t.Errorf("Couldn't read test file %s", tc.yamlPath)
		}
		var ti []TimeInterval
		err = yaml.Unmarshal(f, &ti)
		if err != nil && !tc.expectError {
			t.Errorf("Received unexpected error: %v when parsing %v", err, tc.yamlPath)
		} else if err == nil && tc.expectError {
			t.Errorf("Expected error when unmarshalling %s but didn't receive one", tc.yamlPath)
		} else if !reflect.DeepEqual(ti, tc.intervals) {
			t.Errorf("Error unmarshalling %s: Want %+v, got %+v", tc.yamlPath, tc.intervals, ti)
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
		var tr timeRange
		err := yaml.Unmarshal([]byte(tc.timeString), &tr)
		if err != nil && !tc.expectError {
			t.Errorf("Received unexpected error: %v when parsing %v", err, tc.timeString)
		} else if err == nil && tc.expectError {
			t.Errorf("Expected error for invalid string %s but didn't receive one", tc.timeString)
		} else if !reflect.DeepEqual(tr, tc.timeRange) {
			t.Errorf("Error parsing time string %s: Want %+v, got %+v", tc.timeString, tc.timeRange, tr)
		}
	}
}

func TestParseWeek(t *testing.T) {
	for _, tc := range dayOfWeekStringTestCases {
		var wr []weekdayRange
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
