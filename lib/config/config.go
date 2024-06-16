package config

import (
	"sync"
	"time"
)

type ProtectedIDs struct {
	ids     []int64
	rwMutex sync.RWMutex
}

// scheduledWeeks is the number of weeks for which the barbershop schedule is compiled.
const ScheduledWeeks byte = 26
const NonWorkingDay time.Weekday = time.Monday
const MaxAppointmentBookingMonths int = 1

const (
	TimoutWrite time.Duration = 2 * time.Second
	TimoutRead                = 1 * time.Second
)

// location is the time zone where the barbershop is located.
var Location *time.Location

var Barbers ProtectedIDs

func init() {
	Location = time.FixedZone("MSK", 3*60*60)
}

func InitBarberIDs(ids ...int64) {
	Barbers.ids = ids
}

func (p *ProtectedIDs) IDs() []int64 {
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()
	return p.ids
}

func (p *ProtectedIDs) SetIDs(ids []int64) {
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()
	p.ids = ids
}
