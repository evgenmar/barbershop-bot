package entities

import (
	cfg "barbershop-bot/lib/config"
	tm "barbershop-bot/lib/time"
	"time"
)

type DateRange struct {
	FirstDate time.Time
	LastDate  time.Time
}

// Month returns month for the FirstDate of DateRange.
func (d DateRange) Month() time.Month {
	return d.FirstDate.Month()
}

func (d DateRange) StartWeekday() int {
	startWeekday := int(d.FirstDate.Weekday())
	if startWeekday == 0 {
		return 7
	}
	return startWeekday
}

func (d DateRange) EndWeekday() int {
	endWeekday := int(d.LastDate.Weekday())
	if endWeekday == 0 {
		return 7
	}
	return endWeekday
}

// Month returns the date range for specified month.
// FirstDate and LastDate always sets to first and last days of month correspondingly.
func Month(m tm.Month) DateRange {
	year := int(m/12) + 2001
	month := time.Month(m % 12)
	return DateRange{
		FirstDate: time.Date(year, month, 1, 0, 0, 0, 0, cfg.Location),
		LastDate:  time.Date(year, month+1, 0, 0, 0, 0, 0, cfg.Location),
	}
}

// MonthFromNow returns the date range for the month defined relative to the current date (0 corresponds to the current month, 1 to the next month, and so on).
// FirstDate sets to current date for current month and to first day of month for the rest cases.
// LastDate always sets to last day of month.
func MonthFromNow(deltaMonth byte) DateRange {
	now := time.Now().In(cfg.Location)
	year, month, day := now.Date()
	var firstDate, lastDate time.Time
	if deltaMonth == 0 {
		firstDate = time.Date(year, month, day, 0, 0, 0, 0, cfg.Location)
	} else {
		firstDate = time.Date(year, month+time.Month(deltaMonth), 1, 0, 0, 0, 0, cfg.Location)
	}
	lastDate = time.Date(year, month+time.Month(deltaMonth)+1, 0, 0, 0, 0, 0, cfg.Location)
	return DateRange{
		FirstDate: firstDate,
		LastDate:  lastDate,
	}
}
