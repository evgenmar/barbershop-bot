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
// FirstDate sets to current date for current month and to first day of month for the rest cases.
// LastDate always sets to last day of month.
func Month(m tm.Month) DateRange {
	now := time.Now().In(cfg.Location)
	var day int
	if tm.ParseMonth(now) == m {
		day = now.Day()
	} else {
		day = 1
	}
	year := int(m/12) + 2001
	month := time.Month(m % 12)
	return DateRange{
		FirstDate: time.Date(year, month, day, 0, 0, 0, 0, cfg.Location),
		LastDate:  time.Date(year, month+1, 0, 0, 0, 0, 0, cfg.Location),
	}
}
