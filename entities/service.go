package entities

import (
	tm "barbershop-bot/lib/time"
	"strconv"
)

type Service struct {
	ID         int
	BarberID   int64
	Name       string
	Desciption string
	Price
	Duration tm.Duration
}

type Price uint

func GetPrice(text string) (Price, error) {
	price, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return 0, err
	}
	return Price(price), nil
}

func (p Price) String() string {
	return strconv.FormatUint(uint64(p), 10) + " ₽"
}
