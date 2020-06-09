package timeinterval

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
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

type yamlTimeInterval struct {
	Times       []string `yaml:"times"`
	DaysOfWeek  []string `yaml:"days_of_week"`
	DaysOfMonth []string `yaml:"days_of_month"`
	Months      []string `yaml:"months"`
	Years       []string `yaml:"years"`
}

// TimeLayout specifies the layout to be used in time.Parse() calls for time intervals
const TimeLayout = "15:04"

var validTime string = "^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)"
var validTimeRE *regexp.Regexp = regexp.MustCompile(validTime)

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
			if t.Day() >= validDates.begin && t.Day() <= validDates.end {
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

func FromYaml(in []byte) (TimeInterval, error) {
	y := yamlTimeInterval{}
	ti := TimeInterval{}

	err := yaml.Unmarshal(in, &y)
	if err != nil {
		return TimeInterval{}, err
	}

	if ti.times != nil {
		ti.times = make([]timeRange, len(y.Times))
		for i, timeString := range y.Times {
			time, err := parseTimeString(timeString)
			if err != nil {
				return TimeInterval{}, err
			}
			ti.times[i] = time
		}
	}
	return ti, nil
}

func parseTimeString(in string) (tr timeRange, err error) {
	in = strings.ToLower(in)
	timeStrings := strings.Split(in, "-")
	if len(timeStrings) != 2 {
		return timeRange{}, fmt.Errorf("Couldn't parse time range %s, range specification is badly formatted", in)
	}

	// timeBoundaries will hold the start and end time of the range
	timeBoundaries := make([]int, 2)
	for i, ts := range timeStrings {
		if !validTimeRE.MatchString(ts) {
			return tr, fmt.Errorf("Couldn't parse timestamp %s, invalid format", ts)
		}
		timestampComponents := strings.Split(ts, ":")
		timeStampHours, err := strconv.Atoi(timestampComponents[0])
		if err != nil {
			return tr, err
		}
		timeStampMinutes, err := strconv.Atoi(timestampComponents[1])
		if err != nil {
			return tr, err
		}
		// Timestamps are stored as minutes elapsed in the day, so we must multiply hours by 60
		timeBoundaries[i] = timeStampHours*60 + timeStampMinutes
	}

	beginTime, endTime := timeBoundaries[0], timeBoundaries[1]
	if beginTime == endTime {
		return tr, fmt.Errorf("Couldn't parse time range %s, begin and end times cannot be the same for a time range", in)
	}
	if beginTime > endTime {
		return tr, fmt.Errorf("Couldn't parse time range %s, end time must be greater than begin time", in)
	}

	tr.startMinute = beginTime
	tr.endMinute = endTime
	return tr, nil
}

func parseDayOfWeekString(in string) (tr inclusiveRange, err error) {
	in = strings.ToLower(in)
	if strings.ContainsRune(in, ':') {
		weekdayStrings := strings.Split(in, ":")
		if len(weekdayStrings) != 2 {
			return tr, fmt.Errorf("Coudn't parse day of week range %s, invalid format", in)
		}
	}
	return tr, err
}
