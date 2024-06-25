package sessions

import (
	cfg "barbershop-bot/lib/config"
	"barbershop-bot/lib/e"
	"log"
	"time"
)

type Status struct {
	State State

	//Expiration must use UTC location.
	Expiration time.Time
}

type State byte

const (
	StateStart State = iota + 1
	StateUpdName
	StateUpdPhone
	StateAddBarber
)

var StatusStart Status

func init() {
	expiration, err := time.Parse(time.DateOnly, cfg.InfiniteWorkDate)
	if err != nil {
		log.Fatal(e.Wrap("can't parse default expiration for StatusStart", err))
	}
	StatusStart = Status{
		State:      StateStart,
		Expiration: expiration,
	}
}

// By default new Status lifetime is 24 hours except of StatusStart with a lifetime till 3000-01-01 00:00:00 UTC.
func NewStatus(state State) Status {
	if state == StateStart {
		return StatusStart
	}
	return Status{
		State:      state,
		Expiration: time.Now().In(time.FixedZone("UTC", 0)).Add(24 * time.Hour),
	}
}

func (s Status) isValid() bool {
	return s.Expiration.After(time.Now().In(time.FixedZone("UTC", 0)))
}
