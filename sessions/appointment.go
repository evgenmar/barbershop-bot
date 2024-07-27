package sessions

import (
	tm "barbershop-bot/lib/time"
)

type Appointment struct {
	ID             int
	WorkdayID      int
	ServiceID      int
	Time           tm.Duration
	Duration       tm.Duration
	BarberID       int64
	LastShownMonth tm.Month
}

//TODO: it's better to store appointment info in button's data instead of session
// to avoid errors that occur when accessing multiple menu instances at the same time
