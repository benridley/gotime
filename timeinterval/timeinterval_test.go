package timeinterval

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
			times:      []timeRange{{startMinute: 540, endMinute: 1020}},
			daysOfWeek: []weekdayRange{{begin: 1, end: 5}},
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
			daysOfMonth: []inclusiveRange{{begin: 4, end: 6}},
			months:      []inclusiveRange{{begin: 4, end: 4}},
			years:       []inclusiveRange{{begin: 2020, end: 2020}},
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
			daysOfMonth: []inclusiveRange{{begin: -3, end: -1}},
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
		ranges:      []weekdayRange{{begin: 1, end: 5}, {begin: 6, end: 6}},
		expectError: false,
	},
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