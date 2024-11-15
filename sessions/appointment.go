package sessions

import (
	tm "github.com/evgenmar/barbershop-bot/lib/time"
)

type Appointment struct {
	ID             int
	WorkdayID      int
	ServiceID      int
	Time           tm.Duration
	Duration       tm.Duration
	BarberID       int64
	LastShownMonth tm.Month
	HashStr        string
}
