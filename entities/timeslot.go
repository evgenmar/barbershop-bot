package entities

import tm "barbershop-bot/lib/time"

type Timeslot struct {
	//StartTime is the time interval between the start of the day in local time and the start of the timeslot.
	StartTime tm.Duration
	Duration  tm.Duration
}
