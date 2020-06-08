package timeperiod

import (
	"time"
)

type TimePeriod struct {
	minutesInDay []InclusiveRange
	dates        []InclusiveRange
	months       []InclusiveRange
	days         []InclusiveRange
}

type InclusiveRange struct {
	begin int
	end   int
}

func (tp TimePeriod) ContainsTime(t time.Time) bool {
	if tp.minutesInDay != nil {
		for _, validMinutes := range tp.minutesInDay {
			if t.Minute() >= validMinutes.begin && t.Minute() < validMinutes.end {
				break
			}
			return false
		}
	}
	if tp.dates != nil {
		for _, validDates := range tp.dates {
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
	if tp.days != nil {
		for _, validDays := range tp.days {
			if t.Weekday() >= time.Weekday(validDays.begin) && t.Weekday() <= time.Weekday(validDays.end) {
				break
			}
			return false
		}
	}
	return true
}
