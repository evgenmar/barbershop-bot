package entities

import (
	cfg "barbershop-bot/lib/config"
	"time"
)

type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

// Month returns the date range for the month defined relative to the current date (0 corresponds to the current month, 1 to the next month, and so on).
// StartDate sets to current date for current month and to first day of month for the rest cases.
// EndDate always sets to last day of month.
func Month(m byte) DateRange {
	now := time.Now().In(cfg.Location)
	year, month, day := now.Date()
	var startDate, endDate time.Time
	if m == 0 {
		startDate = time.Date(year, month, day, 0, 0, 0, 0, cfg.Location)
	} else {
		startDate = time.Date(year, month+time.Month(m), 1, 0, 0, 0, 0, cfg.Location)
	}
	endDate = time.Date(year, month+time.Month(m)+1, 0, 0, 0, 0, 0, cfg.Location)
	return DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}
}

// Month returns month for the StartDate of DateRange.
func (d DateRange) Month() time.Month {
	return d.StartDate.Month()
}

func (d DateRange) StartWeekday() int {
	startWeekday := int(d.StartDate.Weekday())
	if startWeekday == 0 {
		return 7
	}
	return startWeekday
}

func (d DateRange) EndWeekday() int {
	endWeekday := int(d.EndDate.Weekday())
	if endWeekday == 0 {
		return 7
	}
	return endWeekday
}
