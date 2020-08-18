package timeinterval

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TimeInterval struct {
	times       []timeRange
	daysOfMonth []inclusiveRange
	months      []inclusiveRange
	daysOfWeek  []weekdayRange
	years       []inclusiveRange
}

/* TimeRange represents a range of minutes within a 1440 minute day, exclusive of the end minute. A day consists of 1440 minutes.
   For example, 5:00PM to end of the day would begin at 1020 and end at 1440. */
type timeRange struct {
	startMinute int
	endMinute   int
}

type weekdayRange struct {
	begin time.Weekday
	end   time.Weekday
}

type inclusiveRange struct {
	begin int
	end   int
}

type yamlTimeRange struct {
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
}

type yamlTimeInterval struct {
	Times       []timeRange    `yaml:"times"`
	DaysOfWeek  []weekdayRange `yaml:"days"`
	DaysOfMonth []string       `yaml:"days_of_month"`
	Months      []string       `yaml:"months"`
	Years       []string       `yaml:"years"`
}

var daysOfWeek = map[string]int{
	"sunday":    0,
	"monday":    1,
	"tuesday":   2,
	"wednesday": 3,
	"thursday":  4,
	"friday":    5,
	"saturday":  6,
}

func (wr *weekdayRange) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dowString string
	if err := unmarshal(&dowString); err != nil {
		return err
	}
	dow, err := parseDayOfWeekString(dowString)
	if err != nil {
		return err
	}
	*wr = dow
	return nil
}

// UnmarshalYAML implements the Unmarshaller interface. Unmarshalling is
// achieved by first unmarshalling into an intermediate type.
func (tr *timeRange) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var y yamlTimeRange
	if err := unmarshal(&y); err != nil {
		return err
	}
	if y.EndTime == "" || y.StartTime == "" {
		return errors.New("Both start and end times must be provided")
	}
	start, err := parseTime(y.StartTime)
	if err != nil {
		return nil
	}
	end, err := parseTime(y.EndTime)
	if err != nil {
		return err
	}
	if start < 0 {
		return errors.New("Start time out of range")
	}
	if end > 1440 {
		return errors.New("End time out of range")
	}
	if start >= end {
		return errors.New("Start time cannot be equal or greater than end time")
	}
	tr.startMinute, tr.endMinute = start, end
	return nil
}

// TimeLayout specifies the layout to be used in time.Parse() calls for time intervals
const TimeLayout = "15:04"

var validTime string = "^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)"
var validTimeRE *regexp.Regexp = regexp.MustCompile(validTime)

// Given a time, determines the number of days in the month that time occurs in.
func daysInMonth(t time.Time) int {
	monthStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	diff := monthEnd.Sub(monthStart)
	return int(diff.Hours() / 24)
}

// ContainsTime returns true if the TimeInterval contains the given time, otherwise returns false
func (tp TimeInterval) ContainsTime(t time.Time) bool {
	if tp.times != nil {
		for _, validMinutes := range tp.times {
			if (t.Hour()*60+t.Minute()) >= validMinutes.startMinute && (t.Hour()*60+t.Minute()) < validMinutes.endMinute {
				break
			}
			return false
		}
	}
	if tp.daysOfMonth != nil {
		for _, validDates := range tp.daysOfMonth {
			var begin, end int
			// Handle negative cases where e.g. -1 refers to the last day of the month
			if validDates.begin < 0 {
				begin = daysInMonth(t) + validDates.begin + 1
			} else {
				begin = validDates.begin
			}
			if validDates.end < 0 {
				end = daysInMonth(t) + validDates.end + 1
			} else {
				end = validDates.end
			}
			if t.Day() >= begin && t.Day() <= end {
				break
			}
			return false
		}
	}
	if tp.months != nil {
		for _, validMonths := range tp.months {
			if t.Month() >= time.Month(validMonths.begin) && t.Month() <= time.Month(validMonths.end) {
				break
			}
			return false
		}
	}
	if tp.daysOfWeek != nil {
		for _, validDays := range tp.daysOfWeek {
			if t.Weekday() >= validDays.begin && t.Weekday() <= validDays.end {
				break
			}
			return false
		}
	}
	if tp.years != nil {
		for _, validYears := range tp.years {
			if t.Year() >= validYears.begin && t.Year() <= validYears.end {
				break
			}
			return false
		}
	}
	return true
}

//func FromYaml(in []byte) (TimeInterval, error) {
//	y := yamlTimeInterval{}
//	ti := TimeInterval{}
//
//	err := yaml.Unmarshal(in, &y)
//	if err != nil {
//		return TimeInterval{}, err
//	}
//
//	if ti.times != nil {
//		ti.times = make([]timeRange, len(y.Times))
//		for i, timeString := range y.Times {
//			time, err := parseTimeString(timeString)
//			if err != nil {
//				return TimeInterval{}, err
//			}
//			ti.times[i] = time
//		}
//	}
//	return ti, nil
//}

// Parses a time into an integer representing minutes elapsed in the day (e.g. 15:23 -> 923)
func parseTime(in string) (mins int, err error) {
	if !validTimeRE.MatchString(in) {
		return 0, fmt.Errorf("Couldn't parse timestamp %s, invalid format", in)
	}
	timestampComponents := strings.Split(in, ":")
	timeStampHours, err := strconv.Atoi(timestampComponents[0])
	if err != nil {
		return 0, err
	}
	timeStampMinutes, err := strconv.Atoi(timestampComponents[1])
	if err != nil {
		return 0, err
	}
	if timeStampHours < 0 || timeStampHours > 24 || timeStampMinutes < 0 || timeStampMinutes > 60 {
		return 0, fmt.Errorf("Timestamp %s out of range", in)
	}
	// Timestamps are stored as minutes elapsed in the day, so multiply hours by 60
	mins = timeStampHours*60 + timeStampMinutes
	return mins, nil
}

func parseDayOfWeekString(in string) (tr weekdayRange, err error) {
	in = strings.ToLower(in)
	if strings.ContainsRune(in, ':') {
		weekdayStrings := strings.Split(in, ":")
		if len(weekdayStrings) != 2 {
			return tr, fmt.Errorf("Coudn't parse day of week range %s, invalid format", in)
		}
		startDay, ok := daysOfWeek[weekdayStrings[0]]
		if !ok {
			return tr, fmt.Errorf("Invalid start day: %s", weekdayStrings[0])
		}
		endDay, ok := daysOfWeek[weekdayStrings[1]]
		if !ok {
			return tr, fmt.Errorf("Invalid end day: %s", weekdayStrings[0])
		}
		if startDay > endDay {
			return tr, fmt.Errorf("Start day cannot be after end day")
		}
		tr = weekdayRange{
			begin: time.Weekday(startDay),
			end:   time.Weekday(endDay),
		}
		return tr, nil
	}
	day, ok := daysOfWeek[in]
	if !ok {
		return tr, fmt.Errorf("Unknown day of week %s", in)
	}
	tr = weekdayRange{
		begin: time.Weekday(day),
		end:   time.Weekday(day),
	}
	return tr, err
}
