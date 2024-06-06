package entities

import (
	tm "barbershop-bot/lib/time"
	"time"
)

type Workday struct {
	BarberID int64

	//Date in local time zone with HH:MM:SS set to 00:00:00.
	Date time.Time

	//StartTime is the time interval between the start of the day in local time and the start of the Workday.
	StartTime tm.Duration

	//EndTime is the time interval between the start of the day in local time and the end of the Workday.
	EndTime tm.Duration
}

func NewWorkday(barberID int64, date time.Time, start, end tm.Duration) Workday {
	return Workday{
		BarberID:  barberID,
		Date:      date,
		StartTime: start,
		EndTime:   end,
	}
}
