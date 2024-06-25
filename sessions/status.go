package sessions

import (
	"time"
)

type Status struct {
	State      State
	Expiration int64
}

type State byte

const (
	StateStart State = iota + 1
	StateUpdName
	StateUpdPhone
	StateAddBarber
)

// By default new Status lifetime is 24 hours except of StatusStart with a lifetime till 3000-01-01 00:00:00 UTC.
func NewStatus(state State) Status {
	if state == StateStart {
		return Status{State: StateStart}
	}
	return Status{
		State:      state,
		Expiration: time.Now().Add(24 * time.Hour).Unix(),
	}
}

func (s Status) isValid() bool {
	if s.State == StateStart {
		return true
	}
	return s.Expiration > time.Now().Unix()
}
