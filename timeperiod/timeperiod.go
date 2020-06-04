package timeperiod

import (
	"time"
)

type TimePeriod struct {
	minutes []InclusiveRange
	hours   []InclusiveRange
	dates   []InclusiveRange
	months  []InclusiveRange
	days    []InclusiveRange
}

type InclusiveRange struct {
	begin int
	end   int
}

func (tp TimePeriod) ContainsTime(t time.Time) bool {
	if tp.minutes != nil {
		for _, validMinutes := range tp.minutes {
			if t.Minute() < validMinutes.begin || t.Minute() > validMinutes.end {
				return false
			}
		}
	}
	if tp.hours != nil {
		for _, validHours := range tp.hours {
			if t.Hour() < validHours.begin || t.Hour() > validHours.end {
				return false
			}
		}
	}
	if tp.dates != nil {
		for _, validDates := range tp.dates {
			if t.Day() < validDates.begin || t.Hour() > validDates.end {
				return false
			}
		}
	}
	if tp.months != nil {
		for _, validMonths := range tp.months {
			if t.Month() < time.Month(validMonths.begin) || t.Month() > time.Month(validMonths.end) {
				return false
			}
		}
	}
	if tp.days != nil {
		for _, validDays := range tp.days {
			if t.Weekday() < time.Weekday(validDays.begin) || t.Weekday() > time.Weekday(validDays.end) {
				return false
			}
		}
	}
	return true
}
