package entities

import (
	tm "github.com/evgenmar/barbershop-bot/lib/time"
)

type Appointment struct {
	ID        int
	UserID    int64
	WorkdayID int
	ServiceID int
	Time      tm.Duration
	Duration  tm.Duration
	Note      string

	//CreatedAt has a format of Unix time
	CreatedAt int64
}
