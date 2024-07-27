package sessions

import (
	"time"
)

type status struct {
	state      State
	expiration int64
}

type State byte

const (
	StateStart State = iota
	StateUpdName
	StateUpdPhone
	StateAddBarber
	StateEnterServiceName
	StateEnterServiceDescription
	StateEnterServicePrice
	StateEditServiceName
	StateEditServiceDescription
	StateEditServicePrice
	StateAddNote
)

// By default new Status lifetime is 24 hours except of StatusStart with a lifetime till 3000-01-01 00:00:00 UTC.
func newStatus(state State) status {
	if state == StateStart {
		return status{state: StateStart}
	}
	return status{
		state:      state,
		expiration: time.Now().Add(24 * time.Hour).Unix(),
	}
}

func (s status) isValid() bool {
	if s.state == StateStart {
		return true
	}
	return s.expiration > time.Now().Unix()
}
