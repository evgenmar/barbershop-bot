package entities

import tm "barbershop-bot/lib/time"

type Service struct {
	ID         int
	BarberID   int64
	Name       string
	Desciption string
	Price      int
	Duration   tm.Duration
}
