package entities

import (
	"errors"
	"fmt"
	"strconv"

	tm "github.com/evgenmar/barbershop-bot/lib/time"
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

func NewPrice(text string) (Price, error) {
	price, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return 0, err
	}
	if price == 0 {
		return 0, errors.New("Price should have a positive value")
	}
	return Price(price), nil
}

func (p Price) String() string {
	return strconv.FormatUint(uint64(p), 10) + " ₽"
}

func (s Service) BtnSignature() string {
	return s.Name + " " + s.Price.String()
}

func (s Service) Info() string {
	return fmt.Sprintf("%s, %s, %s\n%s", s.Name, s.Price.String(), s.Duration.LongString(), s.Desciption)
}

func (s Service) ShortInfo() string {
	return fmt.Sprintf("%s, %s, %s", s.Name, s.Price.String(), s.Duration.LongString())
}
