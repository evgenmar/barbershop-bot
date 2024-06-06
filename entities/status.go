package entities

import (
	"barbershop-bot/lib/e"
	"log"
	"time"
)

type Status struct {
	State State

	//Expiration must ese UTC location.
	Expiration time.Time
}

type State byte

const (
	stateStart State = iota
	stateUpdName
	stateUpdPhone
	stateAddBarber
)

var StatusStart Status

func init() {
	expiration, err := time.Parse(time.DateTime, "3000-01-01 00:00:00")
	if err != nil {
		log.Fatal(e.Wrap("can't parse default expiration for StatusStart", err))
	}
	StatusStart = Status{
		State:      stateStart,
		Expiration: expiration,
	}
}

// By default new Status lifetime is 24 hours except of StatusStart with a lifetime till 3000-01-01 00:00:00 UTC.
func NewStatus(state State) Status {
	if state == stateStart {
		return StatusStart
	}
	return Status{
		State:      state,
		Expiration: time.Now().In(time.FixedZone("UTC", 0)).Add(24 * time.Hour),
	}
}
